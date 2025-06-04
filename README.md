# Orch Metadata Broker Service

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/open-edge-platform/orch-metadata-broker/badge)](https://scorecard.dev/viewer/?uri=github.com/open-edge-platform/orch-metadata-broker)

## Overview

This service is responsible for storing and retrieving metadata, enabling the UI to populate dropdowns with previously entered metadata keys and values. The metadata is stored in a separate file for each project in favor of security.

Read more about Metadata Broker [here](https://github.com/open-edge-platform/orch-metadata-broker/blob/main/docs/metadata-broker.md).

## Get Started

See the [Documentation](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html) to get started using the Metadata Broker.

## Develop

To develop in the Metadata Broker Service, the following development prerequisites are required:

- Go (1.23.8 or later)
- Python with venv (3.1.0 or later)
- [Make](https://www.gnu.org/software/make)
- [Buf](https://github.com/bufbuild/buf)
- A running deployment of the [Edge Management Framework](https://github.com/open-edge-platform/edge-manageability-framework?tab=readme-ov-file)

To build and test Metadata Broker, you can use the following commands:

```shell
# Build the project
make build

# Run the project locally
make run

# If working in a containerized setup, you can use this to build a docker image
make docker-build

# Push the image to a public Amazon ECR registry
make docker-push
```

There are some additional `make` targets to support developer activities:

- `generate` - generates the OpenAPI spec and rest client
- `test` - runs tests for the Go code and rego rules
- `mod-update` - runs both `mod tidy` and `mod vendor`
- `lint` - runs linting on the code

Since all metadata is tied to a project, let's select a project first:

```shell
export PRJ=testProject
```

Now, create a metadata key/value pair:

```shell
curl -X POST \
  -H "Content-Type: application/json" \
  -H "ActiveProjectID: $PRJ" \
  http://localhost:9988/metadata.orchestrator.apis/v1/metadata \
  -d '{
    "metadata": [
      {"key": "color", "value": "red"},
      {"key": "color", "value": "blue"}
    ]
  }'
```

Get all metadata for a project:

```shell
curl -X GET -H "ActiveProjectID: $PRJ" http://localhost:9988/metadata.orchestrator.apis/v1/metadata
```

Delete a specific key/value pair:

```shell
curl -X DELETE -H "ActiveProjectID: $PRJ" http://localhost:9988/metadata.orchestrator.apis/v1/metadata?key=color&value=red
```

Delete all metadata in a project:

```shell
curl -X DELETE http://localhost:9988/metadata.orchestrator.apis/v1/project/$PRJ
```

> Note: This will only delete the project from the Metadata Broker service's file storage. The actual project will still exist in the [Edge Management Framework](https://github.com/open-edge-platform/edge-manageability-framework?tab=readme-ov-file) system.

## Contribute

To learn how to contribute to the project, see the [Contributor's
Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html). The project will accept contributions through Pull-Requests (PRs). PRs must be built successfully by the CI pipeline, pass linter verifications and unit tests.

## Community and Support

To learn more about the project, its community, and governance, visit
the [Edge Orchestrator Community](https://github.com/open-edge-platform).

For support, start with [Troubleshooting](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/troubleshooting/index.html).

## License

The Metadata Broker service is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).

Last Updated Date: April 15, 2025
