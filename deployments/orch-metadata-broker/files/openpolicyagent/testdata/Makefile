# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

OPA          ?= opa
BUNDLE       ?= ../..
PRETTY       ?= -f pretty
TMP_DIR      ?= /tmp
TESTDATA_DIR ?= testdata
TRUE         ?= true
UNDEFINED    ?= undefined

.PHONY: all
all: t1 t2 t3d t3a t4a t4d t5d t5a

t1:
	@# Help: test GetRequest rule as write role  with ActiveProjectId - ALLOWED
	@cat writeRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.GetRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

t2:
	@# Help: test GetRequest rule as read role with ActiveProjectId - ALLOWEDs
	@cat readRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.GetRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

t3d:
	@# Help: test CreateOrUpdate rule as read role - DENIED
	@cat readRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.CreateOrUpdateRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

t3a:
	@# Help: test CreateOrUpdate write as write role - ALLOWED
	@cat writeRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.CreateOrUpdateRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

t4d:
	@# Help: test Delete rule as read role - DENIED
	@cat readRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.DeleteRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

t4a:
	@# Help: test Delete rule as write role - ALLOWED
	@cat writeRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.DeleteRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

t5d:
	@# Help: test DeleteProject rule as read role - DENIED
	@cat readRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.DeleteProjectRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

t5a:
	@# Help: test DeleteProject rule as write role - ALLOWED
	@cat writeRoleWithProject.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.metadatav1.DeleteProjectRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -