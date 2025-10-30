# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.8 as BUILD

WORKDIR /workspace

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./vendor ./vendor
COPY ./api ./api
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux \
   go build -mod=vendor -o /workspace/metadata-service -ldflags="-s -w -extldflags=static" ./cmd/metadata-service/main.go

FROM scratch
COPY --from=build /etc/passwd /etc/passwd
USER nobody
COPY --from=build /workspace/metadata-service /metadata-service
COPY api/spec/openapi.yaml /opt/openapi.yaml

ENTRYPOINT [ "/metadata-service" ]
