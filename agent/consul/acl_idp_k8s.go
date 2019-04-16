package consul

import (
	"fmt"

	k8sidp "github.com/hashicorp/consul/agent/consul/kubernetesidp"
	"github.com/hashicorp/consul/agent/structs"
)

const (
	serviceAccountNamespaceField = "serviceaccount.namespace"
	serviceAccountNameField      = "serviceaccount.name"
	serviceAccountUIDField       = "serviceaccount.uid"
)

type k8sIdentityProviderValidator struct {
	// idp is the underlying state store object with configuration settings
	idp *structs.ACLIdentityProvider

	// tr is used to configure the strategy for doing a token review.
	// Currently the only options are using the kubernetes API or mocking the
	// review. Mocks should only be used in tests.
	tr k8sidp.TokenReviewer

	// publicKeys is an optional list of public key objects used to verify JWTs
	publicKeys []interface{}
}

var _ IdentityProviderValidator = (*k8sIdentityProviderValidator)(nil)

func newK8SIdentityProviderValidator(idp *structs.ACLIdentityProvider) (*k8sIdentityProviderValidator, error) {
	idp = idp.Clone() // avoid monkeying with memdb copy

	if idp.Type != "kubernetes" {
		return nil, fmt.Errorf("%q is not a kubernetes identity provider", idp.Name)
	}

	if idp.KubernetesHost == "" {
		return nil, fmt.Errorf("KubernetesHost is required")
	}

	if idp.KubernetesCACert == "" {
		return nil, fmt.Errorf("KubernetesCACert is required")
	}

	if _, err := k8sidp.ParsePublicKeyPEM([]byte(idp.KubernetesCACert)); err != nil {
		return nil, fmt.Errorf("error parsing kubernetes ca cert: %v", err)
	}

	// This is the bearer token we give the apiserver to use the API.
	if idp.KubernetesServiceAccountJWT == "" {
		return nil, fmt.Errorf("KubernetesServiceAccountJWT is required")
	}
	if err := k8sidp.ValidateJWT(idp.KubernetesServiceAccountJWT); err != nil {
		return nil, fmt.Errorf("KubernetesServiceAccountJWT is not a valid JWT: %v", err)
	}

	// TODO: remove?
	var publicKeys []interface{}
	if len(idp.KubernetesPEMKeys) > 0 {
		publicKeys = make([]interface{}, len(idp.KubernetesPEMKeys))
		for i, cert := range idp.KubernetesPEMKeys {
			data, err := k8sidp.ParsePublicKeyPEM([]byte(cert))
			if err != nil {
				return nil, fmt.Errorf("provided kubernetes public key is invalid: %v", err)
			}
			publicKeys[i] = data
		}
	}

	tr, err := k8sidp.NewTokenReviewer(idp)
	if err != nil {
		return nil, err
	}

	return &k8sIdentityProviderValidator{
		idp:        idp,
		tr:         tr,
		publicKeys: publicKeys,
	}, nil
}

func (v *k8sIdentityProviderValidator) Name() string {
	return v.idp.Name
}

func (v *k8sIdentityProviderValidator) ValidateLogin(loginToken string) (map[string]string, error) {
	sa, err := k8sidp.ParseAndValidateJWT(loginToken, v.publicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate JWT: %v", err)
	}

	// look up the JWT token in the kubernetes API
	r, err := v.tr.Review(loginToken)
	if err != nil {
		return nil, err
	}

	// TODO(rb): is this strictly necessary since the whole point of tokenreview
	// TODO(rb): is that you don't even have to care about the claims?
	// Verify the returned metadata matches the expected data from the service
	// account.
	if sa.EffectiveNamespace() != r.Namespace {
		return nil, fmt.Errorf("JWT namepaces did not match")
	}
	if sa.EffectiveName() != r.OriginalName {
		return nil, fmt.Errorf("JWT names did not match")
	}
	if sa.EffectiveUID() != r.UID {
		return nil, fmt.Errorf("JWT UIDs did not match")
	}

	return map[string]string{
		serviceAccountNamespaceField: r.Namespace,
		serviceAccountNameField:      r.Name, // return the one that could be overridden
		serviceAccountUIDField:       r.UID,
	}, nil
}

func (v *k8sIdentityProviderValidator) MakeFieldMapSelectable(fieldMap map[string]string) interface{} {
	return &k8sFieldDetails{
		ServiceAccount: k8sFieldDetailsServiceAccount{
			Namespace: fieldMap[serviceAccountNamespaceField],
			Name:      fieldMap[serviceAccountNameField],
			UID:       fieldMap[serviceAccountUIDField],
		},
	}
}

type k8sFieldDetails struct {
	ServiceAccount k8sFieldDetailsServiceAccount `bexpr:"serviceaccount"`
}

type k8sFieldDetailsServiceAccount struct {
	Namespace string `bexpr:"namespace"`
	Name      string `bexpr:"name"`
	UID       string `bexpr:"uid"`
}

func (p *k8sIdentityProviderValidator) AvailableFields() []string {
	return []string{
		serviceAccountNamespaceField,
		serviceAccountNameField,
		serviceAccountUIDField,
	}
}
