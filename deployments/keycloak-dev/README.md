<!--
SPDX-FileCopyrightText: 2023-present Intel Corporation

SPDX-License-Identifier: LicenseRef-Intel
-->

# Keycloak Helm Chart configuration for Development

[Keycloak] is Open Source Identity and Access Management for Modern Applications and
Services.

It can also act as a Federated [OpenID Connect] provider. It can connect to a variety of backends.
In this deployment it is not connected to a backend, and just uses its own internal format
persisted to a local Postgres DB.

> This chart can be deployed alongside Application Orch Catalog, App Deployment Manager, Web UI or
> any other microservice that requires an OpenID provider.

## Helm install
Add the Bitnami repo to `helm`, if you don't already have them:
```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

To install the standalone Keycloak server in to a namespace e.g. `orch-platform` use:

```shell
helm -n orch-platform install keycloak bitnami/keycloak --version 16.1.7 -f deployments/keycloak-dev/values.yaml
```

To access this use a port-forward in the cluster
```shell
kubectl -n orch-platform port-forward service/keycloak 8090:80
```

> To test it, browse to http://localhost:8090/realms/master/.well-known/openid-configuration to see the configuration.
>
> Verify the login details at http://localhost:8090/realms/master/account/

See [Authorization](../../docs/authorization.md) for details of how to use with Orchestrator

> Note here that the connection of Application Catalog to keycloak is inside the cluster for the backend services at `http://keycloak`
> whereas the GUI connects to `http://localhost:8090`

## Administration
The Keycloak Admin console can be reached at http://localhost `admin/ChangeMeOn1stLogin!`

## Users
See the `values.yaml` for user names and their default passwords.

## Get Token Directly
To get a token directly for development purposes use:

```shell
curl --location --request POST 'http://localhost:8090/realms/master/protocol/openid-connect/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'grant_type=password' \
--data-urlencode 'client_id=ledge-park-system' \
--data-urlencode 'username=lp-admin-user' \
--data-urlencode 'password=<see password above>' \
--data-urlencode 'scope=openid profile email groups'
```

[Keycloak]: https://www.keycloak.org/
[OpenID Connect]: https://openid.net/connect/
