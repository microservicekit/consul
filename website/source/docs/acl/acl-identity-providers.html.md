---
layout: "docs"
page_title: "ACL Identity Providers"
sidebar_current: "docs-acl-identity-providers"
description: |-
  An ACL Identity Provider is a trusted external party that can be configured to to allow applications to authenticate to Consul using existing credentials and receive ACL Tokens usable within the local datacenter in exchange.
---

-> **1.5.0+:**  This guide only applies in Consul versions 1.5.0 and later.

# ACL Identity Providers

An ACL Identity Provider is a trusted external party that can be configured to
to allow applications to authenticate to Consul using existing credentials and
receive ACL Tokens usable within the local datacenter in exchange.

The only supported type of identity provider in Consul 1.5 is `kubernetes`
but it is expected that more will come later.

## Overview

Without an identity provider integration, a trusted operator needs to be
critically involved in the creation and secure introduction of each ACL Token
to every application that needs one, and ensure that the policies assigned to
these tokens follow the principle of least-privilege.

When running in environments such as a public cloud or when supervised by a
cluster scheduler, applications may already have access to uniquely identifying
credentials that were delivered securely by the platform. Consul identity
provider integrations allow for these credentials to be used to generate ACL
Tokens with properly-scoped policies without operator intervention.

In Consul 1.5 the focus is around simplifying the creation of tokens with the
privileges necessary to participate in a [Connect](/docs/connect/index.html)
service mesh with minimal operator intervention.

## Operator Configuration

An operator needs to configure each identity provider that is to be trusted by
using the API or command line before they can be used by applications.

* **Authentication** - Details about how to authenticate application
  credentials are configured using the `consul acl idp` subcommands or the
  corresponding [API endpoints](/api/acl/identity-providers.html). The specific
  details of configuration are type dependent and described below.

* **Authorization** - One or more Role Binding Rules must be configured
  definiting how to translate trusted identity attributes into privileges
  assigned to the ACL Token that is created. These can be managed with the
  `consul acl rolebindingrule` subcommands or the corresponding [API
  endpoints](/api/acl/role-binding-rules.html).

## Login Process

1. Applications can use the `consul acl login` subcommand or the [login API
   endpoint](/api/acl/acl.html#login-to-identity-provider) to authenticate to
   an identity provider through the Consul leader.

2. The identity provider validates the credentials and returns trusted identity
   attributes to the Consul leader.

3. The Consul leader consults the configured set of Role Binding Rules linked
   to the identity provider to find rules that match the trusted identity
   attributes.

4. If any Role Binding Rules match an ACL Token is created in the local
   datacenter and linked to the computed Roles.

5. Applications can use the `consul acl logout` subcommand or the [logout API
   endpoint](/api/acl/acl.html#logout-from-identity-provider) to destroy their
   token when it is no longer required.

## Kubernetes Identity Provider

The `kubernetes` identity provider type is used to authenticate to Consul using
a Kubernetes Service Account Token. This method of authentication makes it easy
to introduce a Consul token into a Kubernetes Pod.

To use an identity provider of this type the following are required to be
configured:

* **Kubernetes Host** - The address of the Kubernetes API. This should be an
  address that is reachable from all Consul Servers in your datacenter.

* **Kubernetes CA Certificate** - The PEM encoded CA cert for use by the TLS
  client used to talk with the Kubernetes API. NOTE: Every line must end with a
  newline: `\n`

* **Service Account JWT** - A Service Account Token (JWT) used by the Consul
  leader to access the [TokenReview API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#create-tokenreview-v1-authentication-k8s-io)
  and [ServiceAccount API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#read-serviceaccount-v1-core)
  to validate application JWTs during login. 

The following is an example RBAC configuration snippet to grant the necessary
permissions to a service account named `consul-idp`:

```yaml
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: review-tokens
  namespace: default
subjects:
- kind: ServiceAccount
  name: consul-idp
  namespace: default
roleRef:
  kind: ClusterRole
  name: system:auth-delegator
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: service-account-getter
  namespace: default
rules:
- apiGroups: [""]
  resources: ["serviceaccounts"]
  verbs: ["get"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: get-service-accounts
  namespace: default
subjects:
- kind: ServiceAccount
  name: consul-idp
  namespace: default
roleRef:
  kind: ClusterRole
  name: service-account-getter
  apiGroup: rbac.authorization.k8s.io
```

### Kubernetes Login Process

1. The Service Account JWT given to the Consul leader initially accesses the
   TokenReview API to validate the provided JWT is still valid. Kubernetes
   should be running with `--service-account-lookup`. This is defaulted to true
   in Kubernetes 1.7, but any versions prior should ensure the Kubernetes API
   server is started with this setting. 

2. After validating that the provided JWT is still valid, the Consul leader
   looks for an optional annotation of `consul.hashicorp.com/service-name` on
   the resolved service account using the ServiceAccount API.

    ~> TODO: should this be "role-name" or "service-name"?

3. 





### Role Binding Rules

### Login



