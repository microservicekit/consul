package idp

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/mitchellh/mapstructure"
)

type ValidatorFactory func(idp *structs.ACLIdentityProvider) (Validator, error)

type Validator interface {
	// Name returns the name of the identity provider backing this validator.
	Name() string

	// ValidateLogin takes raw user-provided IdP metadata and ensures it is
	// sane, provably correct, and currently valid. Relevant identifying data
	// is extracted and returned for immediate use by the role binding process.
	//
	// Depending upon the provider, it may make sense to use these calls to
	// continue to extend the life of the underlying token.
	//
	// Returns IdP specific metadata suitable for the Role Binding process.
	ValidateLogin(loginToken string) (map[string]string, error)

	// AvailableFields returns a slice of all fields that are returned as a
	// result of ValidateLogin. These are valid fields for use in any
	// BindingRule tied to this identity provider.
	AvailableFields() []string

	// MakeFieldMapSelectable converts a field map as returned by ValidateLogin
	// into a structure suitable for selection with a binding rule.
	MakeFieldMapSelectable(fieldMap map[string]string) interface{}
}

var (
	typesMu sync.RWMutex
	types   = make(map[string]ValidatorFactory)
)

// Register makes an identity provider with the given type available for use.
// If Register is called twice with the same name or if validator is nil, it
// panics.
func Register(name string, factory ValidatorFactory) {
	typesMu.Lock()
	defer typesMu.Unlock()
	if factory == nil {
		panic("idp: Register factory is nil for type " + name)
	}
	if _, dup := types[name]; dup {
		panic("idp: Register called twice for type " + name)
	}
	types[name] = factory
}

func unregisterAllTypes() {
	typesMu.Lock()
	defer typesMu.Unlock()
	// For tests.
	types = make(map[string]ValidatorFactory)
}

func IsRegisteredType(typeName string) bool {
	typesMu.RLock()
	_, ok := types[typeName]
	typesMu.RUnlock()
	return ok
}

// Create instantiates a new Validator for the given identity provider
// configuration.  If no idp is registered with the provided type an error is
// returned.
func Create(idp *structs.ACLIdentityProvider) (Validator, error) {
	typesMu.RLock()
	factory, ok := types[idp.Type]
	typesMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no identity provider registered with type: %s", idp.Type)
	}

	return factory(idp)
}

// Types returns a sorted list of the names of the registered types.
func Types() []string {
	typesMu.RLock()
	defer typesMu.RUnlock()
	var list []string
	for name := range types {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// ParseConfig parses the config block for a identity provider.
func ParseConfig(rawConfig map[string]interface{}, out interface{}) error {
	decodeConf := &mapstructure.DecoderConfig{
		Result:           out,
		WeaklyTypedInput: true,
		ErrorUnused:      true,
	}

	decoder, err := mapstructure.NewDecoder(decodeConf)
	if err != nil {
		return err
	}

	if err := decoder.Decode(rawConfig); err != nil {
		return fmt.Errorf("error decoding config: %s", err)
	}

	return nil
}
