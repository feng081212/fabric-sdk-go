package client

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

type Application struct {
	Capabilities  []string
	Policies      map[string]*Policy
	Organizations []*Organization
	ACLs          map[string]string
}

func DefaultApplication() *Application {
	return &Application{
		Capabilities: []string{"V2_0"},
		Policies:     make(map[string]*Policy),
	}
}

func (p *Application) AddPolicy(name, rule string) {
	if p.Policies == nil {
		p.Policies = make(map[string]*Policy)
	}
	p.Policies[name] = &Policy{ImplicitMetaPolicyType, rule}
}

func (p *Application) BuildConfigGroup() (*common.ConfigGroup, error) {
	configGroup := NewConfigGroup(AdminsPolicyKey)

	if len(p.Capabilities) > 0 {
		_ = addValue(configGroup, AdminsPolicyKey, CapabilitiesKey, CapabilitiesConfigValue(p.Capabilities))
	}

	if err := AddPolicies(configGroup, p.Policies, AdminsPolicyKey); err != nil {
		return nil, errors.Wrapf(err, "error adding policies to application endpoints group ")
	}

	if len(p.ACLs) > 0 {
		aCLs := &peer.ACLs{
			Acls: make(map[string]*peer.APIResource),
		}

		for apiResource, policyRef := range p.ACLs {
			aCLs.Acls[apiResource] = &peer.APIResource{PolicyRef: policyRef}
		}

		_ = addValue(configGroup, AdminsPolicyKey, ACLsKey, aCLs)
	}

	for _, organization := range p.Organizations {
		c, e := organization.BuildConfigGroupForApplication()
		if e != nil {
			return nil, e
		}
		configGroup.Groups[organization.ID] = c
	}

	return configGroup, nil
}
