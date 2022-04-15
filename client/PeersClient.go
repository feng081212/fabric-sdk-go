package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/feng081212/fabric-sdk-go/fabric/endpoints"
	"github.com/golang/protobuf/proto"
	"github.com/feng081212/fabric-protos-go/common"
	"github.com/feng081212/fabric-protos-go/peer"
	"github.com/feng081212/fabric-protos-go/peer/lifecycle"
	"github.com/pkg/errors"
	"sync"
)

type PeersClient struct {
	Peers   []*endpoints.Peer
	Orderer OrdererClient
	Signer  Signer
}

func (p *PeersClient) InvokeChainCode(channelID, chaincodeID string, isInit bool, args [][]byte) (*common.Status, error) {

	request := &endpoints.ChaincodeInvokeRequest{
		ChaincodeID: chaincodeID,
		Args:        args,
		IsInit:      isInit,
	}

	return p.process(context.Background(), channelID, request)
}

func (p *PeersClient) CommitChainCode(channelID string, req *CommitChaincodeRequest) (*common.Status, error) {

	applicationPolicy, e := CreatePolicyBytes(req.SignaturePolicy, req.ChannelConfigPolicy)
	if e != nil {
		return nil, e
	}

	args := &lifecycle.CommitChaincodeDefinitionArgs{
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
		Fcn:         "CommitChaincodeDefinition",
		Args:        [][]byte{ProtoMarshalIgnoreError(args)},
	}

	return p.process(context.Background(), channelID, request)
}

func (p *PeersClient) process(ctx context.Context, channelID string, request *endpoints.ChaincodeInvokeRequest) (*common.Status, error) {

	proposal, header, err := CreateChaincodeInvokeProposal(channelID, p.Signer, request)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateChaincodeInvokeProposal failed")
	}

	responses, _ := p.SendProposal(ctx, proposal)

	if responses == nil || len(responses) == 0 {
		// this should only be empty due to a programming bug
		return nil, errors.New("no proposal responses received")
	}

	var payload []byte
	var endorsements []*peer.Endorsement
	endorsersUsed := make(map[string]bool)

	for _, r := range responses {

		if r == nil || r.ProposalResponse == nil {
			continue
		}

		if r.Response.Status < 200 || r.Response.Status >= 400 {
			//logger.Debugf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
			continue
		}

		proposalResponse := r.ProposalResponse

		if payload == nil {
			payload = proposalResponse.Payload
		}

		if !bytes.Equal(payload, proposalResponse.Payload) {
			return nil, errors.Errorf("ProposalResponsePayloads do not match (base64): '%s' vs '%s'", base64.StdEncoding.EncodeToString(proposalResponse.Payload), base64.StdEncoding.EncodeToString(payload))
		}

		if r.Endorsement != nil {
			key := string(r.Endorsement.Endorser)
			if _, used := endorsersUsed[key]; !used {
				endorsements = append(endorsements, r.Endorsement)
				endorsersUsed[key] = true
			}
		}
	}

	if payload == nil || len(payload) == 0 {
		return nil, errors.New("no payload")
	}

	if len(endorsements) == 0 {
		return nil, errors.New("no endorsements")
	}

	tAction := &peer.TransactionAction{
		Header: header.SignatureHeader,
		Payload: ProtoMarshalIgnoreError(&peer.ChaincodeActionPayload{
			ChaincodeProposalPayload: proposal.Payload,
			Action: &peer.ChaincodeEndorsedAction{
				ProposalResponsePayload: payload, // ChaincodeProposalPayload.TransientMap 不需要已经为空，原SDK中多了一步置空的操作
				Endorsements:            endorsements,
			},
		}),
	}

	return p.Orderer.BroadcastPayload(context.Background(), &common.Payload{
		Header: header,
		Data: ProtoMarshalIgnoreError(&peer.Transaction{
			Actions: []*peer.TransactionAction{tAction},
		}),
	})
}

func (p *PeersClient) SendProposal(ctx context.Context, proposal *peer.Proposal) ([]*endpoints.TransactionProposalResponse, []error) {

	if proposal == nil {
		return nil, []error{errors.New("proposal is required")}
	}

	if len(p.Peers) == 0 {
		return nil, []error{errors.New("targets is required")}
	}

	payloadBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, []error{errors.WithMessage(err, "Marshal Proposal failed")}
	}

	signature, err := p.Signer.Sign(payloadBytes)
	if err != nil {
		return nil, []error{errors.WithMessage(err, "signing of Proposal failed")}
	}

	request := &peer.SignedProposal{
		ProposalBytes: payloadBytes,
		Signature:     signature,
	}

	responses := make([]*endpoints.TransactionProposalResponse, len(p.Peers))
	errs := make([]error, len(p.Peers))

	var wg sync.WaitGroup

	for i, p := range p.Peers {
		wg.Add(1)
		go func(index int, peer *endpoints.Peer) {
			defer wg.Done()
			responses[index], errs[index] = peer.ProcessTransactionProposal(ctx, request)

			if errs[index] != nil {
				//logger.Debugf("Received error response from txn proposal processing: %s", err)
			}
		}(i, p)
	}
	wg.Wait()

	return responses, errs
}
