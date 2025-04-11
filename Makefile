# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

SHELL := bash -eu -o pipefail

BINARY_NAME=metadata-service
BIN_DIR=bin
OAPI_CODEGEN_VERSION		?= v1.12.0


VERSION                 ?= $(shell cat VERSION | tr -d '[:space:]')
PROJECT_NAME       		:= metadata-broker
DOCKER_REGISTRY         ?= 080137407410.dkr.ecr.us-west-2.amazonaws.com
DOCKER_REPOSITORY       ?= edge-orch
DOCKER_SUB_REPOSITORY   ?= orch-ui
GIT_BRANCH 	        ?= $(shell git branch --show-current | sed -r 's/[\/]+/-/g')
DOCKER_IMG_NAME         := $(PROJECT_NAME)
DOCKER_TAG              := $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_SUB_REPOSITORY)/$(DOCKER_IMG_NAME):$(VERSION)
DOCKER_TAG_BRANCH       := $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_SUB_REPOSITORY)/$(DOCKER_IMG_NAME):$(GIT_BRANCH)
HELM_REGISTRY           ?=

## CHART_PREFIX is the prefix of the helm chart
CHART_PREFIX    			?= charts
## CHART_NAME is specified in Chart.yaml
CHART_NAME					?= orch-metadata-broker
## CHART_VERSION is specified in Chart.yaml
CHART_VERSION				?= 0.1.12
## CHART_TEST is specified in test-connection.yaml
CHART_TEST					?= test-connection
## CHART_PATH is given based on repo structure
CHART_PATH					?= "./deployments/${CHART_NAME}"
## CHART_BUILD_DIR is given based on repo structure
CHART_BUILD_DIR				?= ./chart/_output/
## CHART_APP_VERSION is modified on every commit
CHART_APP_VERSION			?= $(VERSION)
## CHART_NAMESPACE can be modified here
CHART_NAMESPACE				?= orch-ui
## CHART_RELEASE can be modified here
CHART_RELEASE				?= orch-metadata-broker
## HELM_REPOSITORY is where we push the helm chart
HELM_REPOSITORY				?=
HELM_REGISTRY				?=

# The endpoint URL of a keycloak server e.g. http://keycloak/realms/master refers to a keycloak server in the cluster
OIDC_SERVER                 ?= http://keycloak.orch-platform.svc/realms/master
# The endpoint URL of a keycloak server e.g. http://localhost:8090/realms/master refers to a keycloak server in the cluster
# by it's externally visible address
OIDC_SERVER_EXTERNAL        ?= http://localhost:8090/realms/master
GOCMD                       := GOPRIVATE="github.com/open-edge-platform/*" go

.PHONY: test

all: build test

# Create the virtualenv with python tools installed
VENV_NAME = venv-mb
$(VENV_NAME): requirements.txt
	echo "Creating virtualenv in $@"
	python3 -m venv $@ ;\
	  . ./$@/bin/activate ; set -u ;\
	  python3 -m pip install --upgrade pip;\
	  python3 -m pip install -r requirements.txt
	echo "To enter virtualenv, run 'source $@/bin/activate'"

license: $(VENV_NAME) ## Check licensing with the reuse tool
	. ./$</bin/activate ; set -u ;\
	reuse --version ;\
	reuse --root . lint

proto-generate: buf-generate openapi-spec-validate
	@# Help: Generate Openapi and validate it

buf-generate:
	@# Help: Generate Openapi Spec and gRPC code
	buf generate

generate: proto-generate rest-client-gen
	@# Help: Generate Openapi and rest client

openapi-spec-validate: $(VENV_NAME)
	@# Help: Runs openapi spec validator
	. ./$</bin/activate ; set -u ;\
	openapi-spec-validator api/spec/openapi.yaml

oapi-codegen:
	@# Help: Install oapi-codegen
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}

rest-client-gen: oapi-codegen
	@# Help: Generate Rest client from the generated openapi spec.
	oapi-codegen -generate client -old-config-style -package restClient -o pkg/restClient/client.go api/spec/openapi.yaml
	oapi-codegen -generate types -old-config-style -package restClient -o pkg/restClient/types.go api/spec/openapi.yaml

build: vendor generate
	$(GOCMD) build -mod=vendor -o ${BIN_DIR}/${BINARY_NAME} ./cmd/metadata-service/main.go

run: build
	PWD=$(shell pwd) ./${BIN_DIR}/${BINARY_NAME} -backupFolder ${PWD}/data -backupFile ${PWD}/metadata.json -openapiSpec ${PWD}/api/spec/openapi.yaml

test:
	# test rego rules
	make -C deployments/orch-metadata-broker/files/openpolicyagent/testdata/ all
	go test ./cmd/...
	go test ./internal/...
	go test ./pkg/...

coverage:
	go test -covermode=count -coverprofile cover.out.tmp \
 		./cmd/... \
 		./internal/... \
 		./pkg/...
 	# remove generated files from coverage report
	cat cover.out.tmp | grep -v "pb.go" | grep -v "pb.gw.go" | grep -v "validate.go" | grep -v "restClient" > cover.out
	go tool cover -html cover.out -o cover.html
	gocover-cobertura < cover.out > coverage.xml

vendor:
	$(GOCMD) mod vendor

docker-build: vendor generate
	docker build -f build/Dockerfile \
	-t $(DOCKER_IMG_NAME):$(GIT_BRANCH) \
	--platform linux/amd64 .

docker-push:
	@# Help: Pushes the docker image
	aws ecr create-repository --region us-west-2 --repository-name  $(DOCKER_REPOSITORY)/$(DOCKER_SUB_REPOSITORY)/$(DOCKER_IMG_NAME) || true
	docker tag $(DOCKER_IMG_NAME):$(GIT_BRANCH) $(DOCKER_TAG_BRANCH)
	docker tag $(DOCKER_IMG_NAME):$(GIT_BRANCH) $(DOCKER_TAG)
	docker push $(DOCKER_TAG)
	docker push $(DOCKER_TAG_BRANCH)

docker-list: ## Print name of docker container image
	@echo "images:"
	@echo "  $(DOCKER_IMG_NAME):"
	@echo "    name: '$(DOCKER_TAG)'"
	@echo "    version: '$(VERSION)'"
	@echo "    gitTagPrefix: 'v'"
	@echo "    buildTarget: 'docker-build'"

chart-clean:
	@# Help: Cleans the build directory of the helm chart
	rm -rf ${CHART_BUILD_DIR}

helm-build: chart-clean apply-version
	@# Help: Builds the helm chart
	helm package \
		--app-version=${CHART_APP_VERSION} \
		--debug \
		--dependency-update \
		--destination ${CHART_BUILD_DIR} \
		${CHART_PATH}

helm-list: ## List helm charts, tag format, and versions in YAML format
	@echo "charts:" ;\
  echo "  $(CHART_NAME):" ;\
  echo -n "    "; grep "^version" "${CHART_PATH}/Chart.yaml"  ;\
  echo "    gitTagPrefix: 'v'" ;\
  echo "    outDir: '${CHART_BUILD_DIR}'" ;\

kind-load:
	@# Help: Load various images into the kind cluster
	kind load docker-image ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

chart-install-kind: docker helm-build keycloak-install-kind
	@# Help: Installs the helm chart in the kind cluster
	helm upgrade --install -n ${CHART_NAMESPACE} ${CHART_RELEASE} \
			--wait --timeout 300s \
			--values ./kind/kind-values.yaml \
			--set openidc.issuer=${OIDC_SERVER} \
 			--set openidc.external=${OIDC_SERVER_EXTERNAL} \
			${CHART_BUILD_DIR}${CHART_NAME}-${CHART_VERSION}.tgz

keycloak-install-kind:
	@# Help: Installs the keycloak server chart in the kind cluster
	@echo "---MAKEFILE KEYCLOAK-INSTALL-KIND---"
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update
	helm install --create-namespace -n ${CHART_NAMESPACE} keycloak bitnami/keycloak --version 16.1.7 -f deployments/keycloak-dev/dev-values.yaml -f deployments/keycloak-dev/values.yaml --timeout 8m
	@echo "---END MAKEFILE KEYCLOAK-INSTALL-KIND---"

chart-uninstall:
	@# Help: Uninstalls the helm chart
	helm uninstall -n ${CHART_NAMESPACE} ${CHART_RELEASE}

chart-test:
	@# Help: Performs smoketest of the deployment
	docker pull appropriate/curl:latest
	docker tag appropriate/curl:latest docker.io/library/appropriate/curl:latest
	kind load docker-image docker.io/library/appropriate/curl:latest
	helm test ${CHART_RELEASE} -n ${CHART_NAMESPACE}
	kubectl -n ${CHART_NAMESPACE} logs ${CHART_RELEASE}-${CHART_TEST} --all-containers | grep orch

chart-test-delete:
	@# Help: Deletes the pod that executed smoketest
	kubectl delete pod ${CHART_RELEASE}-${CHART_NAME}-${CHART_TEST} -n ${CHART_NAMESPACE}

helm-push:
	@# Help: Pushes the helm chart
	aws ecr create-repository --region us-west-2 --repository-name $(DOCKER_REPOSITORY)/$(DOCKER_SUB_REPOSITORY)/$(CHART_PREFIX)/$(CHART_NAME) || true
	helm push ${CHART_BUILD_DIR}${CHART_NAME}-${CHART_APP_VERSION}.tgz oci://$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_SUB_REPOSITORY)/$(CHART_PREFIX)

mod-update:
	@# Help: Update Go modules
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor


lint: lint-go
	@# Help: Runs lint stage

lint-go:
	golangci-lint run --timeout 5m

apply-version:
	yq eval -i '.version = "${VERSION}"' deployments/orch-metadata-broker/Chart.yaml ;
	yq eval -i '.appVersion = "${VERSION}"' deployments/orch-metadata-broker/Chart.yaml ;
