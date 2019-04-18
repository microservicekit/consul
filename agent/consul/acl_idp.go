package consul

import (
	"fmt"

	"github.com/hashicorp/consul/agent/consul/kubeidp"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-bexpr"
)

type IdentityProviderValidator interface {
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

// createIdentityProviderValidator returns an IdentityProviderValidator for the
// given idp configuration.
//
// No caches are updated.
func (s *Server) createIdentityProviderValidator(idp *structs.ACLIdentityProvider) (IdentityProviderValidator, error) {
	switch idp.Type {
	case "kubernetes":
		return kubeidp.NewValidator(idp)
	default:
		return nil, fmt.Errorf("identity provider with name %q found with unknown type %q", idp.Name, idp.Type)
	}
}

type idpValidatorEntry struct {
	Validator   IdentityProviderValidator
	ModifyIndex uint64 // the raft index when this last changed
}

// loadIdentityProviderValidator returns an IdentityProviderValidator for the
// given idp configuration. If the cache is up to date as-of the provided index
// then the cached version is returned, otherwise a new validator is created
// and cached.
func (s *Server) loadIdentityProviderValidator(idx uint64, idp *structs.ACLIdentityProvider) (IdentityProviderValidator, error) {
	if prevIdx, v, ok := s.getCachedIdentityProviderValidator(idp.Name); ok && idx <= prevIdx {
		return v, nil
	}

	v, err := s.createIdentityProviderValidator(idp)

	if err == nil && s.aclIDPValidatorCreateTestHook != nil {
		v, err = s.aclIDPValidatorCreateTestHook(v)
	}

	if err != nil {
		return nil, fmt.Errorf("identity provider validator for %q could not be initialized: %v", idp.Name, err)
	}

	v = s.getOrReplaceIdentityProviderValidator(idp.Name, idx, v)

	return v, nil
}

// getCachedIdentityProviderValidator returns an IdentityProviderValidator for
// the given name exclusively from the cache. If one is not found in the cache
// nil is returned.
func (s *Server) getCachedIdentityProviderValidator(name string) (uint64, IdentityProviderValidator, bool) {
	s.aclIDPValidatorLock.RLock()
	defer s.aclIDPValidatorLock.RUnlock()

	if s.aclIDPValidators != nil {
		v, ok := s.aclIDPValidators[name]
		if ok {
			return v.ModifyIndex, v.Validator, true
		}
	}
	return 0, nil, false
}

// getOrReplaceIdentityProviderValidator updates the cached validator with the
// provided one UNLESS it has been updated by another goroutine in which case
// the updated one is returned.
func (s *Server) getOrReplaceIdentityProviderValidator(name string, idx uint64, v IdentityProviderValidator) IdentityProviderValidator {
	s.aclIDPValidatorLock.Lock()
	defer s.aclIDPValidatorLock.Unlock()

	if s.aclIDPValidators == nil {
		s.aclIDPValidators = make(map[string]*idpValidatorEntry)
	}

	prev, ok := s.aclIDPValidators[name]
	if ok {
		if prev.ModifyIndex >= idx {
			return prev.Validator
		}
	}

	s.logger.Printf("[DEBUG] acl: updating cached identity provider validator for %q", name)

	s.aclIDPValidators[name] = &idpValidatorEntry{
		Validator:   v,
		ModifyIndex: idx,
	}
	return v
}

// purgeIdentityProviderValidators resets the cache of validators.
func (s *Server) purgeIdentityProviderValidators() {
	s.aclIDPValidatorLock.Lock()
	s.aclIDPValidators = make(map[string]*idpValidatorEntry)
	s.aclIDPValidatorLock.Unlock()
}

// evaluateRoleBindings evaluates all current binding rules associated with the
// given identity provider against the verified data returned from the idp
// authentication process.
//
// A list of role links and service identities are returned.
func (s *Server) evaluateRoleBindings(
	validator IdentityProviderValidator,
	verifiedFields map[string]string,
) ([]*structs.ACLServiceIdentity, []structs.ACLTokenRoleLink, error) {
	// Only fetch rules that are relevant for this idp.
	_, rules, err := s.fsm.State().ACLBindingRuleList(nil, validator.Name())
	if err != nil {
		return nil, nil, err
	} else if len(rules) == 0 {
		return nil, nil, nil
	}

	// Convert the fields into something suitable for go-bexpr.
	selectableVars := validator.MakeFieldMapSelectable(verifiedFields)

	// Find all binding rules that match the provided fields.
	var matchingRules []*structs.ACLBindingRule
	for _, rule := range rules {
		if doesBindingRuleMatch(rule, selectableVars) {
			matchingRules = append(matchingRules, rule)
		}
	}
	if len(matchingRules) == 0 {
		return nil, nil, nil
	}

	// For all matching rules compute the attributes of a token.
	var (
		roleLinks         []structs.ACLTokenRoleLink
		serviceIdentities []*structs.ACLServiceIdentity
	)
	for _, rule := range matchingRules {
		bindName, valid, err := computeBindingRuleBindName(rule.BindType, rule.BindName, verifiedFields)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot compute %q bind name for bind target: %v", rule.BindType, err)
		} else if !valid {
			return nil, nil, fmt.Errorf("computed %q bind name for bind target is invalid: %q", rule.BindType, bindName)
		}

		switch rule.BindType {
		case structs.BindingRuleBindTypeService:
			serviceIdentities = append(serviceIdentities, &structs.ACLServiceIdentity{
				ServiceName: bindName,
			})

		case structs.BindingRuleBindTypeRole:
			roleLinks = append(roleLinks, structs.ACLTokenRoleLink{
				Name: bindName,
			})

		default:
			// skip unknown bind type; don't grant privileges
		}
	}

	return serviceIdentities, roleLinks, nil
}

// doesBindingRuleMatch checks that a single binding rule matches the provided
// vars.
func doesBindingRuleMatch(rule *structs.ACLBindingRule, selectableVars interface{}) bool {
	if rule.Selector == "" {
		return true // catch-all
	}

	eval, err := bexpr.CreateEvaluatorForType(rule.Selector, nil, selectableVars)
	if err != nil {
		return false // fails to match if selector is invalid
	}

	result, err := eval.Evaluate(selectableVars)
	if err != nil {
		return false // fails to match if evaluation fails
	}

	return result
}
