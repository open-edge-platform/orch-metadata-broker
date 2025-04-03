# Orch Metadata Broker Service

This service is responsible for storing and retrieving metadata, enabling the UI
to populate dropdowns with previously entered metadata keys and values.

- [Design Documentation](https://github.com/open-edge-platform/orch-metadata-broker/blob/main/docs/metadata-broker.md)

## TL;DR

**Regenerate the gRPC server:**

```shell
make proto-generate
```

**Regenerate the REST client:**

```shell
make rest-client-gen
```

**Run locally:**

```shell
make run
```

**Test:**

```shell
curl -X POST http://localhost:9988/metadata.orchestrator.apis/v1/metadata -d '{
  "metadata": [
    {"key": "customer", "value": "culvers"},
    {"key": "customer", "value": "menards"}
  ]}'
curl -X GET http://localhost:9988/metadata.orchestrator.apis/v1/metadata
curl -X DELETE http://localhost:9988/metadata.orchestrator.apis/v1/metadata?key=customer&value=menards
```

**Test for custom project:**

```shell
export PRJ=testProject
```

```shell
curl -X POST -H "ActiveProjectID: $PRJ" http://localhost:9988/metadata.orchestrator.apis/v1/metadata -d '{
  "metadata": [
    {"key": "color", "value": "red"},
    {"key": "color", "value": "blue"}
  ]}'
curl -X GET -H "ActiveProjectID: $PRJ" http://localhost:9988/metadata.orchestrator.apis/v1/metadata
curl -X DELETE -H "ActiveProjectID: $PRJ" http://localhost:9988/metadata.orchestrator.apis/v1/metadata?key=color&value=red
curl -X DELETE http://localhost:9988/metadata.orchestrator.apis/v1/project/$PRJ
```
