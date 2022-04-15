package client

import (
	"context"
	"fabric-sdk-go/fabric/endpoints"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

type OrdererClient struct {
	Orderer *endpoints.Orderer
	Signer  Signer
	Signers []Signer
}

// GenesisBlock 获取创世区块，区块高度为0
func (p *OrdererClient) GenesisBlock(channelID string) (*common.Block, error) {
	return p.GetBlock(channelID, newSpecificSeekPosition(0))
}

func (p *OrdererClient) GetNewestBlock(channelID string) (*common.Block, error) {
	return p.GetBlock(channelID, newNewestSeekPosition())
}

func (p *OrdererClient) GetConfigBlock(channelID string) (*common.Block, error) {
	block, err := p.GetBlock(channelID, newNewestSeekPosition())
	if err != nil {
		return nil, err
	}
	lc, err := GetLastConfigIndexFromBlock(block)
	if err != nil {
		return nil, err
	}
	if lc == block.Header.Number {
		return block, nil
	}
	return p.GetBlock(channelID, newSpecificSeekPosition(lc))
}

func (p *OrdererClient) GetBlock(channelID string, position *orderer.SeekPosition) (*common.Block, error) {

	payload, err := p.CreatePayload(common.HeaderType_DELIVER_SEEK_INFO, channelID, func() ([]byte, error) {

		seekInfo := &orderer.SeekInfo{
			Start:    position,
			Stop:     position,
			Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
		}

		seekInfoBytes, err := proto.Marshal(seekInfo)
		if err != nil {
			return nil, errors.Wrap(err, "marshal seek info failed")
		}
		return seekInfoBytes, nil
	})

	if err != nil {
		return nil, errors.WithMessage(err, "CreatePayload failed")
	}

	return p.SendPayload(context.Background(), payload)
}

func (p *OrdererClient) DeleteOrganizationalFromConsortium(consortium, mspID string) (*common.Status, error) {
	channelID := "genesis"

	block, err := p.GetConfigBlock(channelID)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannelConfig(channelID, block, func(envelope *common.ConfigEnvelope) error {
		delete(envelope.Config.ChannelGroup.Groups["Consortiums"].Groups[consortium].Groups, mspID)
		return nil
	})
}

func (p *OrdererClient) AddOrganizationalToConsortium(consortium string, organization *Organization) (*common.Status, error) {

	channelID := "genesis"
	orgGroup, err := organization.BuildConfigGroupForConsortium()

	if err != nil {
		return nil, err
	}

	block, err := p.GetConfigBlock(channelID)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannelConfig(channelID, block, func(envelope *common.ConfigEnvelope) error {
		_, b := envelope.Config.ChannelGroup.Groups[ConsortiumsGroupKey].Groups[consortium].Groups[organization.ID]
		if b {
			return fmt.Errorf("consortium[%s] is already exist organizational[%s]", consortium, organization.ID)
		}
		envelope.Config.ChannelGroup.Groups["Consortiums"].Groups[consortium].Groups[organization.ID] = orgGroup
		return nil
	})
}

func (p *OrdererClient) UpdateOrganizationalOfConsortium(consortium, mspID string, updateFunc func(group *common.ConfigGroup) error) (*common.Status, error) {
	channelID := "genesis"

	block, err := p.GetConfigBlock(channelID)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannelConfig(channelID, block, func(envelope *common.ConfigEnvelope) error {
		v, b := envelope.Config.ChannelGroup.Groups[ConsortiumsGroupKey].Groups[consortium].Groups[mspID]
		if !b {
			return fmt.Errorf("consortium[%s] do not have organizational[%s]", consortium, mspID)
		}
		err = updateFunc(v)
		if err != nil {
			return err
		}
		return nil
	})
}

func (p *OrdererClient) CreateChannel(channelID string, channel *Channel) (*common.Status, error) {
	configGroup, e := channel.BuildConfigGroup()
	if e != nil {
		return nil, e
	}

	configGroupTemplate, _ := channel.BuildConfigGroup()
	configGroupTemplate.Groups[ApplicationGroupKey].Values = nil
	configGroupTemplate.Groups[ApplicationGroupKey].Policies = nil

	configUpdate, err := Compute(&common.Config{ChannelGroup: configGroupTemplate}, &common.Config{ChannelGroup: configGroup})
	if err != nil {
		return nil, err
	}
	configUpdate.ChannelId = channelID
	configUpdate.ReadSet.Values[ConsortiumKey] = &common.ConfigValue{Version: 0}
	configUpdate.WriteSet.Values[ConsortiumKey] = &common.ConfigValue{
		Version: 0,
		Value: ProtoMarshalIgnoreError(&common.Consortium{
			Name: channel.ConsortiumName,
		}),
	}

	configBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannel(channelID, configBytes)
}

func (p *OrdererClient) CreateChannelWithBlock(channelID string, chConfigTx []byte) (*common.Status, error) {
	return p.UpdateChannel(channelID, chConfigTx)
}

func (p *OrdererClient) AddOrganizationalToChannel(channelID string, organization *Organization) (*common.Status, error) {

	orgGroup, err := organization.BuildConfigGroupForApplication()

	if err != nil {
		return nil, err
	}

	block, err := p.GetConfigBlock(channelID)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannelConfig(channelID, block, func(envelope *common.ConfigEnvelope) error {
		_, b := envelope.Config.ChannelGroup.Groups[ApplicationGroupKey].Groups[organization.ID]
		if b {
			return fmt.Errorf("channel[%s] is already exist organizational[%s]", channelID, organization.ID)
		}
		envelope.Config.ChannelGroup.Groups[ApplicationGroupKey].Groups[organization.ID] = orgGroup
		return nil
	})
}

func (p *OrdererClient) DeleteOrganizationalToChannel(channelID, mspID string) (*common.Status, error) {

	block, err := p.GetConfigBlock(channelID)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannelConfig(channelID, block, func(envelope *common.ConfigEnvelope) error {
		_, b := envelope.Config.ChannelGroup.Groups[ApplicationGroupKey].Groups[mspID]
		if !b {
			return fmt.Errorf("channel[%s] do not have organizational[%s]", channelID, mspID)
		}
		delete(envelope.Config.ChannelGroup.Groups[ApplicationGroupKey].Groups, mspID)
		return nil
	})
}

func (p *OrdererClient) SetAnchorPeer(mspID, channelID string, anchors ...*AnchorPeer) (*common.Status, error) {
	if len(anchors) == 0 {
		return nil, fmt.Errorf("set anchor peer error to channel[%s] of org[%s]: anchor is nil", channelID, mspID)
	}

	block, e := p.GetConfigBlock(channelID)
	if e != nil {
		return nil, errors.Wrapf(e, "pull config block of channel[%s] error", channelID)
	}

	return p.UpdateChannelConfig(channelID, block, func(configEnvelope *common.ConfigEnvelope) error {
		v, b := configEnvelope.Config.ChannelGroup.Groups[ApplicationGroupKey].Groups[mspID]
		if !b {
			return fmt.Errorf("org[%s] is not in channel[%s]", mspID, channelID)
		}

		var anchorPeers []*peer.AnchorPeer
		for _, anchorPeer := range anchors {
			anchorPeers = append(anchorPeers, &peer.AnchorPeer{
				Host: anchorPeer.Host,
				Port: int32(anchorPeer.Port),
			})
		}
		_ = addValue(v, AdminsPolicyKey, AnchorPeersKey, &peer.AnchorPeers{AnchorPeers: anchorPeers})
		return nil
	})
}

func (p *OrdererClient) SetAnchorPeerWithBlock(channelID string, chConfigTx []byte) (*common.Status, error) {
	return p.UpdateChannel(channelID, chConfigTx)
}

func (p *OrdererClient) UpdateChannelConfig(channelID string, block *common.Block, updateFunc func(*common.ConfigEnvelope) error) (*common.Status, error) {
	envelope := &common.Envelope{}
	if err := proto.Unmarshal(block.Data.Data[0], envelope); err != nil {
		return nil, err
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, err
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, err
	}
	configEnvelopeNew := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelopeNew); err != nil {
		return nil, err
	}
	if err := updateFunc(configEnvelopeNew); err != nil {
		return nil, err
	}

	updateConfig, err := Compute(configEnvelope.Config, configEnvelopeNew.Config)
	if err != nil {
		return nil, err
	}
	updateConfig.ChannelId = channelID

	configBytes, err := proto.Marshal(updateConfig)
	if err != nil {
		return nil, err
	}

	return p.UpdateChannel(channelID, configBytes)
}

func (p *OrdererClient) UpdateChannel(channelID string, updateData []byte) (*common.Status, error) {
	payload, err := p.CreatePayload(common.HeaderType_CONFIG_UPDATE, channelID, func() ([]byte, error) {
		var configSignatures []*common.ConfigSignature
		for _, signer := range p.Signers {
			configSignature, e := CreateConfigSignature(signer, updateData)
			if e != nil {
				return nil, e
			}
			configSignatures = append(configSignatures, configSignature)
		}

		configUpdateEnvelope := &common.ConfigUpdateEnvelope{
			ConfigUpdate: updateData,
			Signatures:   configSignatures,
		}
		configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
		if err != nil {
			return nil, errors.Wrap(err, "marshal configUpdateEnvelope failed")
		}
		return configUpdateEnvelopeBytes, nil
	})

	if err != nil {
		return nil, errors.WithMessage(err, "CreatePayload failed")
	}

	return p.BroadcastPayload(context.Background(), payload)
}

func (p *OrdererClient) CreatePayload(headerType common.HeaderType, channelID string, dataFunc func() ([]byte, error)) (*common.Payload, error) {

	creator, err := p.Signer.Serialize()
	if err != nil {
		return nil, errors.WithMessage(err, "identity from context failed")
	}

	h, err := p.Signer.GetHash()
	if err != nil {
		return nil, errors.WithMessage(err, "hash function creation failed")
	}

	tlsClientCertsHash, err := tlsCertHash(p.Orderer.GetTlsClientCerts())
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get tls cert hash")
	}

	data, err := dataFunc()
	if err != nil {
		return nil, errors.WithMessage(err, "Create Request Data failed")
	}

	return CreatePayload(headerType, channelID, h, creator, data, func(header *common.ChannelHeader) {
		header.TlsCertHash = tlsClientCertsHash
	}), nil
}

func CreateConfigSignature(signer Signer, configtx []byte) (*common.ConfigSignature, error) {

	creator, err := signer.Serialize()
	if err != nil {
		return nil, errors.WithMessage(err, "identity from context failed")
	}

	signatureHeaderBytes := CreateSignatureHeader(creator, nil)

	cfd := ConcatenateBytes(signatureHeaderBytes, configtx)

	signature, err := signer.Sign(cfd)
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
	}

	configSignature := common.ConfigSignature{
		SignatureHeader: signatureHeaderBytes,
		Signature:       signature,
	}
	return &configSignature, nil
}

// BroadcastPayload will send the given payload to some orderer, picking random endpoints
// until all are exhausted
func (p *OrdererClient) BroadcastPayload(context context.Context, payload *common.Payload) (*common.Status, error) {
	// Check if orderers are defined
	if p.Orderer == nil {
		return nil, errors.New("orderer not set")
	}

	envelope, err := signPayload(p.Signer, payload)
	if err != nil {
		return nil, err
	}

	return p.Orderer.SendBroadcast(context, envelope)
}

// SendPayload sends the given payload to each orderer and returns a block response
func (p *OrdererClient) SendPayload(context context.Context, payload *common.Payload) (*common.Block, error) {
	if p.Orderer == nil {
		return nil, errors.New("orderer not set")
	}

	envelope, err := signPayload(p.Signer, payload)
	if err != nil {
		return nil, err
	}

	return p.Orderer.SendDeliver(context, envelope)
}
