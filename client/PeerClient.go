package client

import (
	"context"
	logging "github.com/feng081212/fabric-sdk-go/common/logger"
	"github.com/feng081212/fabric-sdk-go/fabric/endpoints"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/pkg/errors"
	"net/http"
)

var logger = logging.NewLogger("fabsdk/client/peer")

type PeerClient struct {
	Peer   *endpoints.Peer
	Signer Signer
}

func (p *PeerClient) QueryChannels() (*peer.ChannelQueryResponse, error) {

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "GetChannels",
	}

	result := &peer.ChannelQueryResponse{}

	_, _, _, err := p.process(context.Background(), "", request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) JoinChannel(channelID string, ordererClient *OrdererClient) error {

	// 加入通道必须从创世区块开始拉取数据
	block, err := ordererClient.GenesisBlock(channelID)
	if err != nil {
		return errors.WithMessage(err, "missing block input parameter with the required genesis block")
	}

	genesisBlockBytes, err := proto.Marshal(block)
	if err != nil {
		return errors.WithMessage(err, "marshal genesis block failed")
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "JoinChain",
		Args:        [][]byte{genesisBlockBytes},
	}

	_, _, _, err = p.process(context.Background(), "", request, nil)
	return err
}

func (p *PeerClient) GetInstalledChainCodePackageByID(packageID string) (*lifecycle.GetInstalledChaincodePackageResult, error) {

	args := &lifecycle.GetInstalledChaincodePackageArgs{
		PackageId: packageID,
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "GetInstalledChaincodePackage",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.GetInstalledChaincodePackageResult{}

	_, _, _, err := p.process(context.Background(), "", request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) GetInstalledChainCodePackage() (*lifecycle.QueryInstalledChaincodesResult, error) {

	args := &lifecycle.QueryInstalledChaincodesArgs{}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "QueryInstalledChaincodes",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.QueryInstalledChaincodesResult{}

	_, _, _, err := p.process(context.Background(), "", request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) InstallChainCodePackage(pkg []byte) (*lifecycle.InstallChaincodeResult, error) {

	args := &lifecycle.InstallChaincodeArgs{
		ChaincodeInstallPackage: pkg,
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "InstallChaincode",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.InstallChaincodeResult{}

	_, _, _, err := p.process(context.Background(), "", request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) ApproveChainCode(channelID string, req *ApproveChaincodeRequest, ordererClient *OrdererClient) (*common.Status, error) {

	chaincodeSource := &lifecycle.ChaincodeSource{}

	if req.PackageID != "" {
		chaincodeSource.Type = &lifecycle.ChaincodeSource_LocalPackage{
			LocalPackage: &lifecycle.ChaincodeSource_Local{
				PackageId: req.PackageID,
			},
		}
	} else {
		chaincodeSource.Type = &lifecycle.ChaincodeSource_Unavailable_{
			Unavailable: &lifecycle.ChaincodeSource_Unavailable{},
		}
	}

	applicationPolicy, e := CreatePolicyBytes(req.SignaturePolicy, req.ChannelConfigPolicy)
	if e != nil {
		return nil, e
	}

	args := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		ValidationParameter: ProtoMarshalIgnoreError(applicationPolicy),
		InitRequired:        req.InitRequired,
		Collections:         &peer.CollectionConfigPackage{Config: req.CollectionConfig},
		Source:              chaincodeSource,
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "ApproveChaincodeDefinitionForMyOrg",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	ctx := context.Background()

	response, chaincodeInvokeProposal, header, err := p.process(ctx, channelID, request, nil)
	if err != nil {
		return nil, err
	}

	tAction := &peer.TransactionAction{
		Header: header.SignatureHeader,
		Payload: ProtoMarshalIgnoreError(&peer.ChaincodeActionPayload{
			ChaincodeProposalPayload: chaincodeInvokeProposal.Payload,
			Action: &peer.ChaincodeEndorsedAction{
				ProposalResponsePayload: response.Payload, // ChaincodeProposalPayload.TransientMap 不需要已经为空，原SDK中多了一步置空的操作
				Endorsements:            []*peer.Endorsement{response.Endorsement},
			},
		}),
	}

	payload := &common.Payload{
		Header: header,
		Data: ProtoMarshalIgnoreError(&peer.Transaction{
			Actions: []*peer.TransactionAction{tAction},
		}),
	}

	return ordererClient.BroadcastPayload(context.Background(), payload)
}

func (p *PeerClient) QueryApprovedChaincodeDefinition(channelID, ccName string, sequence int64) (*lifecycle.QueryApprovedChaincodeDefinitionResult, error) {

	args := &lifecycle.QueryApprovedChaincodeDefinitionArgs{
		Name:     ccName,
		Sequence: sequence,
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "QueryApprovedChaincodeDefinition",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.QueryApprovedChaincodeDefinitionResult{}

	_, _, _, err := p.process(context.Background(), channelID, request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) CheckCommitReadiness(channelID string, req *CheckChaincodeCommitReadinessRequest) (*lifecycle.CheckCommitReadinessResult, error) {

	applicationPolicy, e := CreatePolicyBytes(req.SignaturePolicy, req.ChannelConfigPolicy)
	if e != nil {
		return nil, e
	}

	args := &lifecycle.CheckCommitReadinessArgs{
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		ValidationParameter: ProtoMarshalIgnoreError(applicationPolicy),
		InitRequired:        req.InitRequired,
		Collections:         &peer.CollectionConfigPackage{Config: req.CollectionConfig},
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "CheckCommitReadiness",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.CheckCommitReadinessResult{}

	_, _, _, err := p.process(context.Background(), channelID, request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) QueryCommitted(channelID, ccName string) (*lifecycle.QueryChaincodeDefinitionResult, error) {

	args := &lifecycle.QueryChaincodeDefinitionArgs{
		Name: ccName,
	}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "QueryChaincodeDefinition",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.QueryChaincodeDefinitionResult{}

	_, _, _, err := p.process(context.Background(), channelID, request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) QueryCommittedOfChannel(channelID string) (*lifecycle.QueryChaincodeDefinitionsResult, error) {

	args := &lifecycle.QueryChaincodeDefinitionsArgs{}

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: "_lifecycle",
		Fcn:         "QueryChaincodeDefinitions",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	result := &lifecycle.QueryChaincodeDefinitionsResult{}

	_, _, _, err := p.process(context.Background(), channelID, request, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PeerClient) QueryChainCode(channelID string, chaincodeID string, isInit bool, args [][]byte) (*peer.Response, error) {

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: chaincodeID,
		Args:        args,
		IsInit:      isInit,
	}

	response, _, _, err := p.process(context.Background(), channelID, request, nil)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

func (p *PeerClient) process(ctx context.Context, channelID string, request *endpoints.ChaincodeInvokeRequest, result proto.Message) (*endpoints.TransactionProposalResponse, *peer.Proposal, *common.Header, error) {

	proposal, header, err := CreateChaincodeInvokeProposal(channelID, p.Signer, request)
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "CreateChaincodeInvokeProposal failed")
	}

	response, err := p.SendProposal(ctx, proposal, result)

	return response, proposal, header, err
}

func (p *PeerClient) SendProposal(ctx context.Context, proposal *peer.Proposal, result proto.Message) (*endpoints.TransactionProposalResponse, error) {

	if proposal == nil {
		return nil, errors.New("proposal is required")
	}

	if p.Peer == nil {
		return nil, errors.New("targets is required")
	}

	payloadBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.WithMessage(err, "Marshal Proposal failed")
	}

	signature, err := p.Signer.Sign(payloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "signing of Proposal failed")
	}

	request := &peer.SignedProposal{
		ProposalBytes: payloadBytes,
		Signature:     signature,
	}

	response, err := p.Peer.ProcessTransactionProposal(ctx, request)
	if err != nil {
		return nil, err
	}
	if response.Status != http.StatusOK {
		return response, errors.Errorf("bad status from %s (%d)", response.Endorser, response.Status)
	}

	if result != nil {
		err = proto.Unmarshal(response.Response.Payload, result)
		if err != nil {
			return response, errors.WithMessage(err, "failed to unmarshal response payload")
		}
	}

	return response, nil
}
