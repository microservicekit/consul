package kubeidp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/agent/structs"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"gopkg.in/square/go-jose.v2/jwt"
	authv1 "k8s.io/api/authentication/v1"
	client_metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	client_authv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	client_corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	client_rest "k8s.io/client-go/rest"
	cert "k8s.io/client-go/util/cert"
)

const (
	serviceAccountNamespaceField = "serviceaccount.namespace"
	serviceAccountNameField      = "serviceaccount.name"
	serviceAccountUIDField       = "serviceaccount.uid"

	serviceAccountServiceNameAnnotation = "consul.hashicorp.com/service-name"
)

// Validator is the wrapper around the relevant portions of the Kubernetes API
// that also conforms to the IdentityProviderValidator interface.
type Validator struct {
	idp      *structs.ACLIdentityProvider
	saGetter client_corev1.ServiceAccountsGetter
	trGetter client_authv1.TokenReviewsGetter
}

func NewValidator(idp *structs.ACLIdentityProvider) (*Validator, error) {
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
	if _, err := cert.ParseCertsPEM([]byte(idp.KubernetesCACert)); err != nil {
		return nil, fmt.Errorf("error parsing kubernetes ca cert: %v", err)
	}

	// This is the bearer token we give the apiserver to use the API.
	if idp.KubernetesServiceAccountJWT == "" {
		return nil, fmt.Errorf("KubernetesServiceAccountJWT is required")
	}
	if _, err := jwt.ParseSigned(idp.KubernetesServiceAccountJWT); err != nil {
		return nil, fmt.Errorf("KubernetesServiceAccountJWT is not a valid JWT: %v", err)
	}

	transport := cleanhttp.DefaultTransport()
	client, err := k8s.NewForConfig(&client_rest.Config{
		Host:        idp.KubernetesHost,
		BearerToken: idp.KubernetesServiceAccountJWT,
		Dial:        transport.DialContext,
		TLSClientConfig: client_rest.TLSClientConfig{
			CAData: []byte(idp.KubernetesCACert),
		},
		ContentConfig: client_rest.ContentConfig{
			ContentType: "application/json",
		},
	})
	if err != nil {
		return nil, err
	}

	return &Validator{
		idp:      idp,
		saGetter: client.CoreV1(),
		trGetter: client.AuthenticationV1(),
	}, nil
}

func (v *Validator) Name() string {
	return v.idp.Name
}

func (v *Validator) ValidateLogin(loginToken string) (map[string]string, error) {
	if _, err := jwt.ParseSigned(loginToken); err != nil {
		return nil, fmt.Errorf("failed to parse and validate JWT: %v", err)
	}

	// Check TokenReview for the bulk of the work.
	trResp, err := v.trGetter.TokenReviews().Create(&authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token: loginToken,
		},
	})

	if err != nil {
		return nil, err
	} else if trResp.Status.Error != "" {
		return nil, fmt.Errorf("lookup failed: %s", trResp.Status.Error)
	}

	if !trResp.Status.Authenticated {
		return nil, errors.New("lookup failed: service account jwt not valid")
	}

	// The username is of format: system:serviceaccount:(NAMESPACE):(SERVICEACCOUNT)
	parts := strings.Split(trResp.Status.User.Username, ":")
	if len(parts) != 4 {
		return nil, errors.New("lookup failed: unexpected username format")
	}

	// Validate the user that comes back from token review is a service account
	if parts[0] != "system" || parts[1] != "serviceaccount" {
		return nil, errors.New("lookup failed: username returned is not a service account")
	}

	var (
		saNamespace = parts[2]
		saName      = parts[3]
		saUID       = string(trResp.Status.User.UID)
	)

	// Check to see  if there is an override name on the ServiceAccount object.
	sa, err := v.saGetter.ServiceAccounts(saNamespace).Get(saName, client_metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("annotation lookup failed: %v", err)
	}

	annotations := sa.GetObjectMeta().GetAnnotations()
	if serviceNameOverride, ok := annotations[serviceAccountServiceNameAnnotation]; ok {
		saName = serviceNameOverride
	}

	return map[string]string{
		serviceAccountNamespaceField: saNamespace,
		serviceAccountNameField:      saName,
		serviceAccountUIDField:       saUID,
	}, nil
}

func (p *Validator) AvailableFields() []string {
	return []string{
		serviceAccountNamespaceField,
		serviceAccountNameField,
		serviceAccountUIDField,
	}
}

func (v *Validator) MakeFieldMapSelectable(fieldMap map[string]string) interface{} {
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
