package client

import (
	"fmt"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/hasher"
	"github.com/feng081212/fabric-sdk-go/fabric/policies"
	"github.com/feng081212/fabric-protos-go/common"
	"github.com/pkg/errors"
	"math"
	"strconv"
)

type Consortium struct {
	Name             string
	Capabilities     []string
	Policies         map[string]*Policy
	Organizations    []*Organization
	OrdererEndpoints *OrdererEndpoints
}

func DefaultConsortium(name string) *Consortium {
	return &Consortium{
		Capabilities: []string{"V2_0"},
		Policies:     make(map[string]*Policy),
		Name:         name,
	}
}

func (p *Consortium) AddPolicy(name, rule string) {
	if p.Policies == nil {
		p.Policies = make(map[string]*Policy)
	}
	p.Policies[name] = &Policy{ImplicitMetaPolicyType, rule}
}

func (p *Consortium) GenesisBlock() (*common.Block, error) {

	configGroup, e := p.BuildConfigGroup()
	if e != nil {
		return nil, e
	}

	return CreateGenesisBlock(ConfigChannelName, configGroup)
}

func newConsortiumsConfigGroup(consortiumName string, organizations []*Organization) (*common.ConfigGroup, error) {
	consortiumConfigGroup := NewConfigGroup(ordererAdminsPolicyName)

	pp, _ := NewImplicitMetaPolicy(&common.ImplicitMetaPolicy{
		Rule:      common.ImplicitMetaPolicy_ANY,
		SubPolicy: AdminsPolicyKey,
	})
	_ = addValue(consortiumConfigGroup, ordererAdminsPolicyName, ChannelCreationPolicyKey, pp)

	for _, organization := range organizations {
		c, e := organization.BuildConfigGroupForConsortium()
		if e != nil {
			return nil, e
		}
		consortiumConfigGroup.Groups[organization.ID] = c
	}

	configGroup := NewConfigGroup(ordererAdminsPolicyName)
	configGroup.Groups[consortiumName] = consortiumConfigGroup

	ppp, _ := NewSignaturePolicy(policies.AcceptAllPolicy)
	addPolicy(configGroup, ordererAdminsPolicyName, AdminsPolicyKey, ppp)

	return configGroup, nil
}

func (p *Consortium) BuildConfigGroup() (*common.ConfigGroup, error) {

	if p.Name == "" {
		return nil, fmt.Errorf("consortium name can not be nil")
	}

	consortiumGroup, e := newConsortiumsConfigGroup(p.Name, p.Organizations)
	if e != nil {
		return nil, e
	}

	var addresses []string
	if p.OrdererEndpoints != nil {
		for _, o := range p.OrdererEndpoints.Orderers {
			address := o.Host + ":" + strconv.Itoa(int(o.Port))
			addresses = append(addresses, address)
		}
	}

	configGroup := NewConfigGroup(AdminsPolicyKey)

	_ = addValue(configGroup, AdminsPolicyKey, HashingAlgorithmKey, &common.HashingAlgorithm{Name: string(hasher.SHA256)})
	_ = addValue(configGroup, AdminsPolicyKey, BlockDataHashingStructureKey, &common.BlockDataHashingStructure{Width: math.MaxUint32})

	if e = AddPolicies(configGroup, p.Policies, AdminsPolicyKey); e != nil {
		return nil, errors.Wrapf(e, "error adding policies to consortium '%s'", p.Name)
	}

	if len(p.Capabilities) > 0 {
		_ = addValue(configGroup, AdminsPolicyKey, CapabilitiesKey, CapabilitiesConfigValue(p.Capabilities))
	}

	if p.OrdererEndpoints != nil {
		if configGroup.Groups[OrdererGroupKey], e = p.OrdererEndpoints.BuildConfigGroup(); e != nil {
			return nil, e
		}
	}

	if addresses != nil && len(addresses) > 0 {
		_ = addValue(configGroup, ordererAdminsPolicyName, OrdererAddressesKey, &common.OrdererAddresses{Addresses: addresses})
	}

	configGroup.Groups[ConsortiumsGroupKey] = consortiumGroup

	return configGroup, nil
}
