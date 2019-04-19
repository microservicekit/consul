package consul

import (
	"fmt"

	idp_pkg "github.com/hashicorp/consul/agent/consul/idp"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-bexpr"

	// register this as a builtin idp
	_ "github.com/hashicorp/consul/agent/consul/idp/k8s"
)

type idpValidatorEntry struct {
	Validator   idp_pkg.Validator
	ModifyIndex uint64 // the raft index when this last changed
}

// loadIdentityProviderValidator returns an idp_pkg.Validator for the
// given idp configuration. If the cache is up to date as-of the provided index
// then the cached version is returned, otherwise a new validator is created
// and cached.
func (s *Server) loadIdentityProviderValidator(idx uint64, idp *structs.ACLIdentityProvider) (idp_pkg.Validator, error) {
	if prevIdx, v, ok := s.getCachedIdentityProviderValidator(idp.Name); ok && idx <= prevIdx {
		return v, nil
	}

	v, err := idp_pkg.Create(idp)
	if err != nil {
		return nil, fmt.Errorf("identity provider validator for %q could not be initialized: %v", idp.Name, err)
	}

	v = s.getOrReplaceIdentityProviderValidator(idp.Name, idx, v)

	return v, nil
}

// getCachedIdentityProviderValidator returns an IdentityProviderValidator for
// the given name exclusively from the cache. If one is not found in the cache
// nil is returned.
func (s *Server) getCachedIdentityProviderValidator(name string) (uint64, idp_pkg.Validator, bool) {
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
func (s *Server) getOrReplaceIdentityProviderValidator(name string, idx uint64, v idp_pkg.Validator) idp_pkg.Validator {
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
	validator idp_pkg.Validator,
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
