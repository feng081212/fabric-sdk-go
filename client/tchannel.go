package client

import (
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/hasher"
	"github.com/feng081212/fabric-protos-go/common"
	"github.com/pkg/errors"
	"math"
)

type Channel struct {
	ChannelID      string
	ConsortiumName string
	Capabilities   []string
	Policies       map[string]*Policy
	Organizations  []*Organization
	Application    *Application
}

func DefaultChannel(consortiumName, channelID string) *Channel {
	return &Channel{
		Capabilities:   []string{"V2_0"},
		Policies:       make(map[string]*Policy),
		ChannelID:      channelID,
		ConsortiumName: consortiumName,
	}
}

func (p *Channel) AddPolicy(name, rule string) {
	if p.Policies == nil {
		p.Policies = make(map[string]*Policy)
	}
	p.Policies[name] = &Policy{ImplicitMetaPolicyType, rule}
}

func (p *Channel) GenesisBlock() (*common.Block, error) {
	configGroup, e := p.BuildConfigGroup()
	if e != nil {
		return nil, e
	}

	return CreateGenesisBlock(p.ChannelID, configGroup)
}

func (p *Channel) BuildConfigGroup() (*common.ConfigGroup, error) {

	if p.Application == nil {
		return nil, errors.New("cannot define a new channel with no Application section")
	}

	if p.ConsortiumName == "" {
		return nil, errors.New("cannot define a new channel with no Consortium value")
	}

	configGroup := NewConfigGroup(AdminsPolicyKey)

	_ = addValue(configGroup, AdminsPolicyKey, HashingAlgorithmKey, &common.HashingAlgorithm{Name: string(hasher.SHA256)})
	_ = addValue(configGroup, AdminsPolicyKey, BlockDataHashingStructureKey, &common.BlockDataHashingStructure{Width: math.MaxUint32})

	if e := AddPolicies(configGroup, p.Policies, AdminsPolicyKey); e != nil {
		return nil, errors.Wrapf(e, "error adding policies to channel '%s'", p.ChannelID)
	}

	if len(p.Capabilities) > 0 {
		_ = addValue(configGroup, AdminsPolicyKey, CapabilitiesKey, CapabilitiesConfigValue(p.Capabilities))
	}

	_ = addValue(configGroup, AdminsPolicyKey, ConsortiumKey, &common.Consortium{Name: p.ConsortiumName})

	if p.Application != nil {
		var e error
		if configGroup.Groups[ApplicationGroupKey], e = p.Application.BuildConfigGroup(); e != nil {
			return nil, e
		}
	}

	return configGroup, nil
}
