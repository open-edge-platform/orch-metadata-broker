<!--
SPDX-FileCopyrightText: 2023-present Intel Corporation

SPDX-License-Identifier: LicenseRef-Intel
-->

# Keycloak Helm Chart configuration for Development

[Keycloak] is an Open Source Identity and Access Management solution
for modern applications and services.

It can also act as a Federated [OpenID Connect] provider,
connecting to a variety of backends. In this deployment,
it is not connected to a backend and uses its own
internal format persisted to a local Postgres DB.

> This chart can be deployed alongside the Application Orch Catalog,
> App Deployment Manager, Web UI, or any other microservice
> that requires an OpenID provider.

## Helm Install

Add the Bitnami repo to `helm`, if you don't already have it:

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

To install the standalone Keycloak server into a namespace,
e.g., `orch-platform`, use:

```shell
helm -n orch-platform install keycloak bitnami/keycloak --version 16.1.7 -f deployments/keycloak-dev/values.yaml
```

To access this, use a port-forward in the cluster:

```shell
kubectl -n orch-platform port-forward service/keycloak 8090:80
```

> To test it, browse to
> `http://localhost:8090/realms/master/.well-known/openid-configuration`
> to see the configuration.
>
> Verify the login details at `http://localhost:8090/realms/master/account/`
>
> Note that the connection of the Application Catalog to Keycloak
> is inside the cluster for the backend services at
> `http://keycloak`, whereas the GUI connects to `http://localhost:8090`.

## Administration

The Keycloak Admin console can be reached at `http://localhost`.
Credentials: `admin/ChangeMeOn1stLogin!`

## Users

See the `values.yaml` for usernames and their default passwords.

## Get Token Directly

To get a token directly for development purposes, use:

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
