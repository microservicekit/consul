package k8s

import (
	"testing"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/stretchr/testify/require"
)

func TestValidateLogin(t *testing.T) {
	testSrv := StartTestAPIServer(t)
	defer testSrv.Stop()

	testSrv.AuthorizeJWT(goodJWT_A)
	testSrv.SetAllowedServiceAccount(
		"default",
		"demo",
		"76091af4-4b56-11e9-ac4b-708b11801cbe",
		"",
		goodJWT_B,
	)

	idp := &structs.ACLIdentityProvider{
		Name:        "test-k8s",
		Description: "k8s test",
		Type:        "kubernetes",
		Config: map[string]interface{}{
			"Host":              testSrv.Addr(),
			"CACert":            testSrv.CACert(),
			"ServiceAccountJWT": goodJWT_A,
		},
	}
	validator, err := NewValidator(idp)
	require.NoError(t, err)

	t.Run("invalid idp token", func(t *testing.T) {
		_, err := validator.ValidateLogin("invalid")
		require.Error(t, err)
	})

	t.Run("valid idp token", func(t *testing.T) {
		fields, err := validator.ValidateLogin(goodJWT_B)
		require.NoError(t, err)
		require.Equal(t, map[string]string{
			"serviceaccount.namespace": "default",
			"serviceaccount.name":      "demo",
			"serviceaccount.uid":       "76091af4-4b56-11e9-ac4b-708b11801cbe",
		}, fields)
	})

	// annotate the account
	testSrv.SetAllowedServiceAccount(
		"default",
		"demo",
		"76091af4-4b56-11e9-ac4b-708b11801cbe",
		"alternate-name",
		goodJWT_B,
	)

	t.Run("valid idp token with annotation", func(t *testing.T) {
		fields, err := validator.ValidateLogin(goodJWT_B)
		require.NoError(t, err)
		require.Equal(t, map[string]string{
			"serviceaccount.namespace": "default",
			"serviceaccount.name":      "alternate-name",
			"serviceaccount.uid":       "76091af4-4b56-11e9-ac4b-708b11801cbe",
		}, fields)
	})
}

func TestNewValidator(t *testing.T) {
	ca := connect.TestCA(t, nil)

	type IDP = *structs.ACLIdentityProvider

	makeIDP := func(f func(idp IDP)) *structs.ACLIdentityProvider {
		idp := &structs.ACLIdentityProvider{
			Name:        "test-k8s",
			Description: "k8s test",
			Type:        "kubernetes",
			Config: map[string]interface{}{
				"Host":              "https://abc:8443",
				"CACert":            ca.RootCert,
				"ServiceAccountJWT": goodJWT_A,
			},
		}
		if f != nil {
			f(idp)
		}
		return idp
	}

	for _, test := range []struct {
		name string
		idp  *structs.ACLIdentityProvider
		ok   bool
	}{
		// bad
		{"wrong type", makeIDP(func(idp IDP) {
			idp.Type = "invalid"
		}), false},
		{"extra config", makeIDP(func(idp IDP) {
			idp.Config["extra"] = "config"
		}), false},
		{"wrong type of config", makeIDP(func(idp IDP) {
			idp.Config["Host"] = []int{12345}
		}), false},
		{"missing host", makeIDP(func(idp IDP) {
			delete(idp.Config, "Host")
		}), false},
		{"missing ca cert", makeIDP(func(idp IDP) {
			delete(idp.Config, "CACert")
		}), false},
		{"invalid ca cert", makeIDP(func(idp IDP) {
			idp.Config["CACert"] = "invalid"
		}), false},
		{"invalid jwt", makeIDP(func(idp IDP) {
			idp.Config["ServiceAccountJWT"] = "invalid"
		}), false},
		{"garbage host", makeIDP(func(idp IDP) {
			idp.Config["Host"] = "://:12345"
		}), false},
		// good
		{"normal", makeIDP(nil), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			v, err := NewValidator(test.idp)
			if test.ok {
				require.NoError(t, err)
				require.NotNil(t, v)
			} else {
				require.NotNil(t, err)
				require.Nil(t, v)
			}
		})
	}
}

// 'default/consul-idp-token-review-account-token-m62ds'
const goodJWT_A = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImNvbnN1bC1pZHAtdG9rZW4tcmV2aWV3LWFjY291bnQtdG9rZW4tbTYyZHMiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiY29uc3VsLWlkcC10b2tlbi1yZXZpZXctYWNjb3VudCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6Ijc1ZTNjYmVhLTRiNTYtMTFlOS1hYzRiLTcwOGIxMTgwMWNiZSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmNvbnN1bC1pZHAtdG9rZW4tcmV2aWV3LWFjY291bnQifQ.uMb66tZ8d8gNzS8EnjlkzbrGKc5M-BESwS5B46IUbKfdMtajsCwgBXICytWKQ2X7wfm4QQykHVaElijBlO8QVvYeYzQE0uy75eH9EXNXmRh862YL_Qcy_doPC0R6FQXZW99S5Joc-3riKsq7N-sjEDBshOqyfDaGfan3hxaiV4Bv4hXXWRFUQ9aTAfPVvk1FQi21U9Fbml9ufk8kkk6gAmIEA_o7p-ve6WIhm48t7MJv314YhyVqXdrvmRykPdMwj4TfwSn3pTJ82P4NgSbXMJhwNkwIadJPZrM8EfN5ISpR4EW3jzP3IHtgQxrIovWQ9TQib1Z5zdRaLWaFVm6XaQ"

// 'default/demo'
const goodJWT_B = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlbW8tdG9rZW4ta21iOW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVtbyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6Ijc2MDkxYWY0LTRiNTYtMTFlOS1hYzRiLTcwOGIxMTgwMWNiZSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlbW8ifQ.ZiAHjijBAOsKdum0Aix6lgtkLkGo9_Tu87dWQ5Zfwnn3r2FejEWDAnftTft1MqqnMzivZ9Wyyki5ZjQRmTAtnMPJuHC-iivqY4Wh4S6QWCJ1SivBv5tMZR79t5t8mE7R1-OHwst46spru1pps9wt9jsA04d3LpV0eeKYgdPTVaQKklxTm397kIMUugA6yINIBQ3Rh8eQqBgNwEmL4iqyYubzHLVkGkoP9MJikFI05vfRiHtYr-piXz6JFDzXMQj9rW6xtMmrBSn79ChbyvC5nz-Nj2rJPnHsb_0rDUbmXY5PpnMhBpdSH-CbZ4j8jsiib6DtaGJhVZeEQ1GjsFAZwQ"

// 'default/monolith'
const goodJWT_C = "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Im1vbm9saXRoLXRva2VuLWp6amQ2Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Im1vbm9saXRoIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiY2IyZTAyM2UtNjA3YS0xMWU5LWIxOWUtNDhlNmM4YjhlY2I1Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OmRlZmF1bHQ6bW9ub2xpdGgifQ.b-Q3M-DDdhTkxY4GCKU7tZIbGdB07ASY7D0K5ci5omTRrjqBw73M4CQU3g2yWAv5Lkb9koUvMcnFNc9PpoqXBFweZg9z82sBcUFX2QmQGy4uIOY6qkIVqLjer_c26-lGvdlAnidkrAOXqNPrc-Iqcfo1Qc4pqnLMOQLDuEGnjPcVoQzB3kPVchwOG8r3uhYfnqfmmilRk88IwTCjmZL-YEnFXEkrypCFGAVZvK992CY0zEonEC3Dr19-72HpC0U6xWuqi6nprX4__S-phd708u43drHNGpff84BkAuImbSTGvpEU5oJ8_Z27swi2DBI-bozKLFEebgx-Y53BAMtXtA"
