package client

import (
	"fabric-sdk-go/fabric/crypto/certs"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/orderer/etcdraft"
	"github.com/pkg/errors"
)

type OrdererEndpoint struct {
	Host          string
	Port          uint32
	ClientTlsCert string
	ServerTlsCert string
}

func (p *OrdererEndpoint) GetRaftConsenter() (*etcdraft.Consenter, error) {
	clientCert, e := certs.ReadPemFromString(p.ClientTlsCert)
	if e != nil {
		return nil, e
	}
	serverCert, e := certs.ReadPemFromString(p.ServerTlsCert)
	if e != nil {
		return nil, e
	}
	return &etcdraft.Consenter{
		Host:          p.Host,
		Port:          p.Port,
		ClientTlsCert: clientCert,
		ServerTlsCert: serverCert,
	}, nil
}

func DefaultOrdererEndpoints() *OrdererEndpoints {
	return &OrdererEndpoints{
		BatchTimeout: &orderer.BatchTimeout{Timeout: "2s"},
		BatchSize: &orderer.BatchSize{
			MaxMessageCount:   500,
			AbsoluteMaxBytes:  10 * 1024 * 1024,
			PreferredMaxBytes: 2 * 1024 * 1024,
		},
		ChannelRestrictions: &orderer.ChannelRestrictions{
			MaxCount: 0,
		},
		Capabilities: []string{"V2_0"},
		Policies:     make(map[string]*Policy),
	}
}

type OrdererEndpoints struct {
	Organizations       []*Organization
	Orderers            []*OrdererEndpoint
	Policies            map[string]*Policy
	Capabilities        []string
	BatchSize           *orderer.BatchSize
	BatchTimeout        *orderer.BatchTimeout
	ChannelRestrictions *orderer.ChannelRestrictions
}

func (p *OrdererEndpoints) AddPolicy(name, rule string) {
	if p.Policies == nil {
		p.Policies = make(map[string]*Policy)
	}
	p.Policies[name] = &Policy{ImplicitMetaPolicyType, rule}
}

func (p *OrdererEndpoints) AddOrderer(ordererEndpoint ...*OrdererEndpoint) {
	if len(ordererEndpoint) == 0 {
		return
	}
	p.Orderers = append(p.Orderers, ordererEndpoint...)
}

func (p *OrdererEndpoints) AddOrganization(organizations ...*Organization) {
	if len(organizations) == 0 {
		return
	}
	p.Organizations = append(p.Organizations, organizations...)
}

func (p *OrdererEndpoints) BuildConfigGroup() (*common.ConfigGroup, error) {
	configGroup := NewConfigGroup(AdminsPolicyKey)

	_ = addValue(configGroup, AdminsPolicyKey, BatchSizeKey, p.BatchSize)
	_ = addValue(configGroup, AdminsPolicyKey, BatchTimeoutKey, p.BatchTimeout)
	_ = addValue(configGroup, AdminsPolicyKey, ChannelRestrictionsKey, p.ChannelRestrictions)
	if len(p.Capabilities) > 0 {
		_ = addValue(configGroup, AdminsPolicyKey, CapabilitiesKey, CapabilitiesConfigValue(p.Capabilities))
	}

	consensusType, e := GetConsensusType(p)
	if e != nil {
		return nil, e
	}

	_ = addValue(configGroup, AdminsPolicyKey, ConsensusTypeKey, consensusType)

	for _, organization := range p.Organizations {
		c, e := organization.BuildConfigGroupForOrderer()
		if e != nil {
			return nil, e
		}
		configGroup.Groups[organization.ID] = c
	}

	if err := AddPolicies(configGroup, p.Policies, AdminsPolicyKey); err != nil {
		return nil, errors.Wrapf(err, "error adding policies to orderer endpoints group ")
	}

	return configGroup, nil
}

func GetConsensusType(ordererEndpoints *OrdererEndpoints) (*orderer.ConsensusType, error) {
	raftConfigMetadata := &etcdraft.ConfigMetadata{
		Options: &etcdraft.Options{
			TickInterval:         "500ms",
			ElectionTick:         10,
			HeartbeatTick:        1,
			MaxInflightBlocks:    5,
			SnapshotIntervalSize: 16 * 1024 * 1024, // 16 MB
		},
	}
	for _, o := range ordererEndpoints.Orderers {
		consenter, e := o.GetRaftConsenter()
		if e != nil {
			return nil, e
		}
		raftConfigMetadata.Consenters = append(raftConfigMetadata.Consenters, consenter)
	}

	consensusType := &orderer.ConsensusType{
		Type:     ConsensusTypeEtcdRaft,
		Metadata: ProtoMarshalIgnoreError(raftConfigMetadata),
	}

	return consensusType, nil
}
