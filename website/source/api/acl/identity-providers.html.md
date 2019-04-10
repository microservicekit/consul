---
layout: api
page_title: ACL Identity Providers - HTTP API
sidebar_current: api-acl-identity-providers
description: |-
  The /acl/idp endpoints manage Consul's ACL Identity Providers.
---

-> **1.5.0+:**  The APIs are available in Consul versions 1.5.0 and later. The documentation for the legacy ACL API is [here](/api/acl/legacy.html)

# ACL Identity Provider HTTP API

The `/acl/idp` endpoints [create](#create-an-identity-provider),
[read](#read-an-identity-provider), [update](#update-an-identity-provider),
[list](#list-identity-providers) and [delete](#delete-an-identity-provider)
ACL identity providers in Consul.  For more information about ACLs, please see
the [ACL Guide](/docs/guides/acl.html).

## Create an Identity Provider

This endpoint creates a new ACL identity provider.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/acl/idp`                   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `NO`             | `none`            | `none`        | `acl:write`  |

### Parameters

- `Name` `(string: <required>)` - Specifies a name for the ACL identity
  provider. The name can only contain alphanumeric characters as well as `-`
  and `_` and must be unique. This field is immutable.
   
~> TODO(rb): update this.

- `Description` `(string: "")` - Free form human readable description of the identity provider.

- `Type` `(string: <required>)` - The type of identity provider being
  configured.  The only allowed value in Consul 1.5.0 is `"kubernetes"`. This
  field is immutable.

#### Parameters for Type=kubernetes

~> TODO: link to idp subsection about kubernetes for details

- `KubernetesHost` `(string: <required>)` - Must be a host string, a host:port
  pair, or a URL to the base of the Kubernetes API server. 

- `KubernetesCACert` `(string: <required>)` - PEM encoded CA cert for use by
  the TLS client used to talk with the Kubernetes API. NOTE: Every line must
  end with a newline (`\n`).

- `KubernetesServiceAccountJWT` `(string: <required>)` A service account JWT
  used to access the TokenReview API to validate other JWTs during login. It
  also must be able to read ServiceAccount annotations. 

~> TODO: link to idp subsection about kubernetes for details

### Sample Payload

```json
{
    "Name": "minikube",
    "Description": "dev minikube cluster",
    "Type": "kubernetes",
    "KubernetesHost": "https://192.0.2.42:8443",
    "KubernetesCACert": "-----BEGIN CERTIFICATE-----\n...-----END CERTIFICATE-----\n",
    "KubernetesServiceAccountJWT": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9..."
}
```

### Sample Request

```sh
$ curl -X PUT \
    --data @payload.json \
    http://127.0.0.1:8500/v1/acl/idp
```

### Sample Response

```json
{
    "Name": "minikube",
    "Description": "dev minikube cluster",
    "Type": "kubernetes",
    "KubernetesHost": "https://192.0.2.42:8443",
    "KubernetesCACert": "-----BEGIN CERTIFICATE-----\n...-----END CERTIFICATE-----\n",
    "KubernetesServiceAccountJWT": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9...",
    "CreateIndex": 15,
    "ModifyIndex": 15
}
```

## Read an Identity Provider

This endpoint reads an ACL identity provider with the given name. If no
identity provider exists with the given name, a 404 is returned instead of a
200 response.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/acl/idp/:name`             | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `YES`            | `all`             | `none`        | `acl:read`   |

### Parameters

- `name` `(string: <required>)` - Specifies the name of the ACL identity
  provider to read. This is required and is specified as part of the URL path.

### Sample Request

```sh
$ curl -X GET http://127.0.0.1:8500/v1/acl/idp/minikube
```

### Sample Response

```json
{
    "Name": "minikube",
    "Description": "dev minikube cluster",
    "Type": "kubernetes",
    "KubernetesHost": "https://192.0.2.42:8443",
    "KubernetesCACert": "-----BEGIN CERTIFICATE-----\n...-----END CERTIFICATE-----\n",
    "KubernetesServiceAccountJWT": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9...",
    "CreateIndex": 15,
    "ModifyIndex": 224
}
```

## Update an Identity Provider

This endpoint updates an existing ACL identity provider.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/acl/idp/:name`             | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `NO`             | `none`            | `none`        | `acl:write`  |

### Parameters

- `Name` `(string: <required>)` - Specifies the name of the identity provider to update. This is
   required in the URL path but may also be specified in the JSON body. If specified
   in both places then they must match exactly.

- `Description` `(string: "")` - Free form human readable description of the identity provider.

- `Type` `(string: <required>)` - Specifies the type of the identity provider
  being updated.  This field is immutable so if present in the body then it
  must match the existing value. If not present then the value will be filled
  in by Consul.

#### Parameters for Type=kubernetes

~> TODO: link to idp subsection about kubernetes for details

- `KubernetesHost` `(string: <required>)` - Must be a host string, a host:port
  pair, or a URL to the base of the Kubernetes API server. 

- `KubernetesCACert` `(string: <required>)` - PEM encoded CA cert for use by
  the TLS client used to talk with the Kubernetes API. NOTE: Every line must
  end with a newline (`\n`).

- `KubernetesServiceAccountJWT` `(string: <required>)` A service account JWT
  used to access the TokenReview API to validate other JWTs during login. It
  also must be able to read ServiceAccount annotations. 

~> TODO: link to idp subsection about kubernetes for details

### Sample Payload

```json
{
    "Name": "minikube",
    "Description": "updated name",
    "KubernetesHost": "https://192.0.2.42:8443",
    "KubernetesCACert": "-----BEGIN CERTIFICATE-----\n...-----END CERTIFICATE-----\n",
    "KubernetesServiceAccountJWT": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9..."
}
```

### Sample Request

```sh
$ curl -X PUT \
    --data @payload.json \
    http://127.0.0.1:8500/v1/acl/idp/minikube
```

### Sample Response

```json
{
    "Name": "minikube",
    "Description": "updated name",
    "Type": "kubernetes",
    "KubernetesHost": "https://192.0.2.42:8443",
    "KubernetesCACert": "-----BEGIN CERTIFICATE-----\n...-----END CERTIFICATE-----\n",
    "KubernetesServiceAccountJWT": "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9...",
    "CreateIndex": 15,
    "ModifyIndex": 224
}
```

## Delete an Identity Provider

This endpoint deletes an ACL identity provider.

~> Deleting an identity provider will also immediately delete all associated
[role binding rules](/api/acl/role-binding-rules.html) as well as marking any
outstanding [tokens](/api/acl/tokens.html) created from this identity provider
as eligible for deletion.

| Method   | Path                      | Produces                   |
| -------- | ------------------------- | -------------------------- |
| `DELETE` | `/acl/idp/:name`          | `application/json`         |

Even though the return type is application/json, the value is either true or
false indicating whether the delete succeeded.

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `NO`             | `none`            | `none`        | `acl:write`  |

### Parameters

- `name` `(string: <required>)` - Specifies the name of the ACL identity
  provider to delete. This is required and is specified as part of the URL
  path.

### Sample Request

```sh
$ curl -X DELETE \
    http://127.0.0.1:8500/v1/acl/idp/minikube
```

### Sample Response
```json
true
```

## List Identity Providers

This endpoint lists all the ACL identity providers.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/acl/idps`                 | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `YES`            | `all`             | `none`        | `acl:read`   |

## Sample Request

```sh
$ curl -X GET http://127.0.0.1:8500/v1/acl/idps
```

### Sample Response

-> **Note** - The `KubernetesCACert` and `KubernetesServiceAccountJWT` fields
are not included in the listing and must be retrieved by the
[identity provider reading endpoint](#read-an-identity-provider).

```json
[
    {
        "Name": "minikube-1",
        "Description": "",
        "Type": "kubernetes",
        "KubernetesHost": "https://192.0.2.42:8443",
        "CreateIndex": 14,
        "ModifyIndex": 14
    },
    {
        "Name": "minikube-2",
        "Description": "",
        "Type": "kubernetes",
        "KubernetesHost": "https://192.0.2.43:8443",
        "CreateIndex": 15,
        "ModifyIndex": 15
    }
]
```
