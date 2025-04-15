# Orch Metadata Broker Service

## Overview

This service is responsible for storing and retrieving metadata, enabling the UI to populate dropdowns with previously entered metadata keys and values. The metadata is stored in a separate file for each project in favor of security.

Finish the section by sending them to a relevant section of the
documentation:

Read more about Metadata Broker [here](https://github.com/open-edge-platform/orch-metadata-broker/blob/main/docs/metadata-broker.md).

## Get Started

See the [Documentation](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html) to get started using the Metadata Broker.

## Develop

To develop in the Metadata Broker Service, the following development prerequisites are required:

- Go (1.23.8 or later)
- Python (3.1.0 or later)
- Make
- [Buf] (https://github.com/bufbuild/buf)
- A running deployment of the [Edge Management Framework](https://github.com/open-edge-platform/edge-manageability-framework?tab=readme-ov-file)

To build and test Metadata Broker, you can use the following commands:
```shell
# Regenerate the  OpenApi spec and rest client
make generate

# Build the project
make build

# Run the project locally
make run
```

## Contribute

To learn how to contribute to the project, see the [Contributor's
Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html). The project will accept contributions through Pull-Requests (PRs). PRs must be built successfully by the CI pipeline, pass linters verifications and the unit tests.

## Community and Support

To learn more about the project, its community, and governance, visit
the \[Edge Orchestrator Community\](https://website-name.com).

For support, start with [Troubleshooting](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/troubleshooting/index.html).

## License

The Metadata Broker service is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).

Last Updated Date: April 15, 2025