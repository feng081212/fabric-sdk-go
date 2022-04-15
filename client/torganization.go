package client

import (
	"fmt"
	"github.com/feng081212/fabric-protos-go/common"
	"github.com/feng081212/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

func DefaultOrganization(mspID string) *Organization {
	if mspID == "" {
		return nil
	}
	organization := &Organization{
		ID:       mspID,
		Name:     mspID,
		Policies: make(map[string]*Policy),
	}
	rule := fmt.Sprintf("OR('%s.admin', '%s.peer', '%s.client', '%s.orderer')", mspID, mspID, mspID, mspID)
	organization.AddPolicy(ReadersPolicyKey, rule)
	organization.AddPolicy(WritersPolicyKey, rule)
	organization.AddPolicy(AdminsPolicyKey, rule)
	organization.AddPolicy(EndorsementPolicyKey, rule)
	organization.AddPolicy(BlockValidationPolicyKey, "OR('all.orderer')")
	return organization
}

// Organization encodes the organization-level configuration needed in config transactions.
type Organization struct {
	Name             string
	ID               string
	MspConfig        *MspConfig
	MSPType          string
	Policies         map[string]*Policy
	AnchorPeers      []*AnchorPeer
	OrdererEndpoints []string
}

func (p *Organization) AddPolicy(name, rule string) {
	if p.Policies == nil {
		p.Policies = make(map[string]*Policy)
	}
	p.Policies[name] = &Policy{SignaturePolicyType, rule}
}

func (p *Organization) BuildConfigGroup() (*common.ConfigGroup, error) {
	if p.ID != p.MspConfig.MspID {
		return nil, fmt.Errorf("msp id error: organization.ID[%s] != organization.MspConfig.MspID[%s]", p.ID, p.MspConfig.MspID)
	}

	mspConf, err := p.MspConfig.BuildMspConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "1 - Error loading MSP configuration for org: %s", p.ID)
	}

	orgConfigGroup := NewConfigGroup(AdminsPolicyKey)

	if err := AddPolicies(orgConfigGroup, p.Policies, AdminsPolicyKey); err != nil {
		return nil, errors.Wrapf(err, "error adding policies to orderer org group '%s'", p.Name)
	}

	_ = addValue(orgConfigGroup, AdminsPolicyKey, MSPKey, mspConf)

	if p.OrdererEndpoints != nil && len(p.OrdererEndpoints) > 0 {
		_ = addValue(orgConfigGroup, AdminsPolicyKey, EndpointsKey, &common.OrdererAddresses{Addresses: p.OrdererEndpoints})
	}

	if p.AnchorPeers != nil && len(p.AnchorPeers) > 0 {
		var anchors []*peer.AnchorPeer
		for _, anchorPeer := range p.AnchorPeers {
			anchors = append(anchors, &peer.AnchorPeer{
				Host: anchorPeer.Host,
				Port: int32(anchorPeer.Port),
			})
		}
		_ = addValue(orgConfigGroup, AdminsPolicyKey, AnchorPeersKey, &peer.AnchorPeers{AnchorPeers: anchors})
	}

	return orgConfigGroup, nil
}

func (p *Organization) BuildConfigGroupForConsortium() (*common.ConfigGroup, error) {
	configGroup, e := p.BuildConfigGroup()
	if e != nil {
		return nil, e
	}
	// 添加机构到联盟中，Endpoints AnchorPeers 值不需要，传了该值无法添加成功
	delete(configGroup.Values, AnchorPeersKey)
	delete(configGroup.Values, EndpointsKey)
	return configGroup, nil
}

func (p *Organization) BuildConfigGroupForOrderer() (*common.ConfigGroup, error) {
	configGroup, e := p.BuildConfigGroup()
	if e != nil {
		return nil, e
	}
	// 添加机构到Orderer中，AnchorPeers 值不需要，传了该值无法添加成功
	delete(configGroup.Values, AnchorPeersKey)
	return configGroup, nil
}

func (p *Organization) BuildConfigGroupForApplication() (*common.ConfigGroup, error) {
	configGroup, e := p.BuildConfigGroup()
	if e != nil {
		return nil, e
	}
	// 添加机构到通道中，Endpoints 值不需要，传了该值无法添加成功
	delete(configGroup.Values, EndpointsKey)
	return configGroup, nil
}
