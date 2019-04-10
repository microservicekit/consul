---
layout: api
page_title: ACL Role Binding Rules - HTTP API
sidebar_current: api-acl-role-binding-rules
description: |-
  The /acl/rolebindingrule endpoints manage Consul's ACL Role Binding Rules.
---

-> **1.5.0+:** The APIs are available in Consul versions 1.5.0 and later. The documentation for the legacy ACL API is [here](/api/acl/legacy.html)

# ACL Role Binding Rule HTTP API

The `/acl/rolebindingrule` endpoints [create](#create-a-role-binding-rule),
[read](#read-a-role-binding-rule), [update](#update-a-role-binding-rule),
[list](#list-roles) and [delete](#delete-a-role-binding-rule)  ACL role binding
rules in Consul.  For more information about ACLs, please see the 
[ACL Guide](/docs/guides/acl.html).

## Create a Role Binding Rule

This endpoint creates a new ACL role binding rule.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/acl/rolebindingrule`       | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `NO`             | `none`            | `none`        | `acl:write`  |

### Parameters

- `Description` `(string: "")` - Free form human readable description of the role binding rule.

- `IDPName` `(string: <required>)` - The name of the identity provider that
  this rule applies to. This field is immutable.

- `RoleName` `(string: <required>)` - The name of a role to bind to. Can
  optionally use `{{ field_name }}` template syntax to generate a name using available
  identity provider fields. Available fields are documented in the
  [Identity Provider Guide](/docs/guides/acl-identity-providers.html).

- `MustExist` `(bool: false)` - If true, indicates that at login time the named
  role must already exist for this role binding rule to apply. This is described
  in more detail in the [Identity Provider Guide](/docs/guides/acl-identity-providers.html).

- `Matches` `(array<Match>)` - The list of match selectors. Individual matches
  logically are used as a disjunction (OR) when matching identities presented.
  If unset or empty the role binding rule will match all identities in the 
  configured identity provider.

  - `Selector` `(array<string>)` - A non-empty list of field selectors.
    Elements logically are used as a conjunction (AND) when matching identities
    presented. The syntax of each element is of the form `field=value` and
    available fields are documented in the 
    [Identity Provider Guide](/docs/guides/acl-identity-providers.html).
  
### Sample Payload

```json
{
    "Description": "example rule",
    "IDPName": "minikube",
    "Matches": [
        {
            "Selector": [
                "serviceaccount.namespace=default"
            ]
        },
        {
            "Selector": [
                "serviceaccount.namespace=dev",
                "serviceaccount.name=demo"
            ]
        }
    ],
    "RoleName": "{{ serviceaccount.name }}"
}
```

### Sample Request

```sh
$ curl -X PUT \
    --data @payload.json \
    http://127.0.0.1:8500/v1/acl/rolebindingrule
```

### Sample Response

```json
{
    "ID": "000ed53c-e2d3-e7e6-31a5-c19bc3518a3d",
    "Description": "example rule",
    "IDPName": "minikube",
    "Matches": [
        {
            "Selector": [
                "serviceaccount.namespace=default"
            ]
        },
        {
            "Selector": [
                "serviceaccount.namespace=dev",
                "serviceaccount.name=demo"
            ]
        }
    ],
    "RoleName": "{{ serviceaccount.name }}",
    "CreateIndex": 17,
    "ModifyIndex": 17
}
```

## Read a Role Binding Rule

This endpoint reads an ACL role binding rule with the given ID. If no role
binding rule exists with the given ID, a 404 is returned instead of a 200
response.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/acl/rolebindingrule/:id`   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `YES`            | `all`             | `none`        | `acl:read`   |

### Parameters

- `id` `(string: <required>)` - Specifies the UUID of the ACL role binding rule
  to read. This is required and is specified as part of the URL path.

### Sample Request

```sh
$ curl -X GET http://127.0.0.1:8500/v1/acl/rolebindingrule/000ed53c-e2d3-e7e6-31a5-c19bc3518a3d
```

### Sample Response

```json
{
    "ID": "000ed53c-e2d3-e7e6-31a5-c19bc3518a3d",
    "Description": "example rule",
    "IDPName": "minikube",
    "Matches": [
        {
            "Selector": [
                "serviceaccount.namespace=default"
            ]
        },
        {
            "Selector": [
                "serviceaccount.namespace=dev",
                "serviceaccount.name=demo"
            ]
        }
    ],
    "RoleName": "{{ serviceaccount.name }}",
    "CreateIndex": 17,
    "ModifyIndex": 17
}
```

## Update a Role Binding Rule

This endpoint updates an existing ACL role binding rule.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `PUT`  | `/acl/rolebindingrule/:id`   | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `NO`             | `none`            | `none`        | `acl:write`  |

### Parameters

- `ID` `(string: <required>)` - Specifies the ID of the role binding rule to update. This is
   required in the URL path but may also be specified in the JSON body. If specified
   in both places then they must match exactly.

- `Description` `(string: "")` - Free form human readable description of the role binding rule.

- `IDPName` `(string: <required>)` - Specifies the name of the identity
  provider that this rule applies to. This field is immutable so if present in
  the body then it must match the existing value. If not present then the value
  will be filled in by Consul.

- `RoleName` `(string: <required>)` - The name of a role to bind to. Can
  optionally use `{{ field_name }}` template syntax to generate a name using available
  identity provider fields. Available fields are documented in the
  [Identity Provider Guide](/docs/guides/acl-identity-providers.html).

- `MustExist` `(bool: false)` - If true, indicates that at login time the named
  role must already exist for this role binding rule to apply. This is described
  in more detail in the [Identity Provider Guide](/docs/guides/acl-identity-providers.html).

- `Matches` `(array<Match>)` - The list of match selectors. Individual matches
  logically are used as a disjunction (OR) when matching identities presented.
  If unset or empty the role binding rule will match all identities in the 
  configured identity provider.

  - `Selector` `(array<string>)` - A non-empty list of field selectors.
    Elements logically are used as a conjunction (AND) when matching identities
    presented. The syntax of each element is of the form `field=value` and
    available fields are documented in the 
    [Identity Provider Guide](/docs/guides/acl-identity-providers.html).
  

### Sample Payload

```json
{
    "Description": "updated rule",
    "Matches": [
        {
            "Selector": [
                "serviceaccount.namespace=default"
            ]
        }
    ],
    "RoleName": "k8s-{{ serviceaccount.name }}"
}
```

### Sample Request

```sh
$ curl -X PUT \
    --data @payload.json \
    http://127.0.0.1:8500/v1/acl/rolebindingrule/000ed53c-e2d3-e7e6-31a5-c19bc3518a3d
```

### Sample Response

```json
{
    "ID": "000ed53c-e2d3-e7e6-31a5-c19bc3518a3d",
    "Description": "updated rule",
    "IDPName": "minikube",
    "Matches": [
        {
            "Selector": [
                "serviceaccount.namespace=default"
            ]
        }
    ],
    "RoleName": "k8s-{{ serviceaccount.name }}",
    "CreateIndex": 17,
    "ModifyIndex": 18
}
```

## Delete a Role Binding Rule

This endpoint deletes an ACL role binding rule.

| Method   | Path                      | Produces                   |
| -------- | ------------------------- | -------------------------- |
| `DELETE` | `/acl/rolebindingrule/:id`| `application/json`         |

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

- `id` `(string: <required>)` - Specifies the UUID of the ACL role binding rule
  to delete. This is required and is specified as part of the URL path.

### Sample Request

```sh
$ curl -X DELETE \
    http://127.0.0.1:8500/v1/acl/rolebindingrule/000ed53c-e2d3-e7e6-31a5-c19bc3518a3d
```

### Sample Response
```json
true
```

## List Role Binding Rules

This endpoint lists all the ACL role binding rules.

| Method | Path                         | Produces                   |
| ------ | ---------------------------- | -------------------------- |
| `GET`  | `/acl/rolebindingrules`                 | `application/json`         |

The table below shows this endpoint's support for
[blocking queries](/api/index.html#blocking-queries),
[consistency modes](/api/index.html#consistency-modes),
[agent caching](/api/index.html#agent-caching), and
[required ACLs](/api/index.html#acls).

| Blocking Queries | Consistency Modes | Agent Caching | ACL Required |
| ---------------- | ----------------- | ------------- | ------------ |
| `YES`            | `all`             | `none`        | `acl:read`   |

## Parameters

- `idp` `(string: "")` - Filters the role binding rule list to those role
  binding rules that are linked with the specific named identity provider.

## Sample Request

```sh
$ curl -X GET http://127.0.0.1:8500/v1/acl/rolebindingrules
```

### Sample Response

```json
[
    {
        "ID": "000ed53c-e2d3-e7e6-31a5-c19bc3518a3d",
        "Description": "example 1",
        "IDPName": "minikube-1",
        "RoleName": "k8s-{{ serviceaccount.name }}",
        "CreateIndex": 17,
        "ModifyIndex": 17
    },
    {
        "ID": "b4f0a0a3-69f2-7a4f-6bef-326034ace9fa",
        "Description": "example 2",
        "IDPName": "minikube-2",
        "Matches": [
            {
                "Selector": [
                    "serviceaccount.namespace=default"
                ]
            }
        ],
        "RoleName": "k8s-{{ serviceaccount.name }}",
        "CreateIndex": 18,
        "ModifyIndex": 18
    }
]
```
