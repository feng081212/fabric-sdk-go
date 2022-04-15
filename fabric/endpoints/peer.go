/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoints

import (
	"context"
	"crypto/tls"
	"github.com/feng081212/fabric-sdk-go/common/utils"
	"github.com/feng081212/fabric-sdk-go/common/utils/grpcutils"
	certs2 "github.com/feng081212/fabric-sdk-go/fabric/crypto/certs"
	retry2 "github.com/feng081212/fabric-sdk-go/fabric/errors/retry"
	status2 "github.com/feng081212/fabric-sdk-go/fabric/errors/status"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"regexp"
	"time"

	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var chaincodeNotFoundPattern = regexp.MustCompile(`(chaincode [^ ]+ not found)|(could not find chaincode with name)|(cannot get package for chaincode)`)

// Peer represents a node in the target blockchain network to which
// HFC sends endorsement proposals, transaction ordering or query requests.
type Peer struct {
	serverName      string
	mspID           string
	url             string
	target          string
	keepaliveParams *keepalive.ClientParameters
	failFast        bool
	inSecure        bool
	properties      map[string]interface{}
	timeout         time.Duration
	grpcOpts        []grpc.DialOption
	tlsCaCerts      *x509.CertPool
	tlsClientCerts  []tls.Certificate
	retryOpts       retry2.Opts
}

// MSPID gets the Peer mspID.
func (p *Peer) MSPID() string {
	return p.mspID
}

// URL gets the Peer URL. Required property for the instance objects.
// It returns the address of the Peer.
func (p *Peer) URL() string {
	return p.url
}

// Properties returns the properties of a peer.
func (p *Peer) Properties() map[string]interface{} {
	return p.properties
}

func (p *Peer) String() string {
	return p.url
}

func (p *Peer) GetTimeout() time.Duration {
	if p.timeout != 0 {
		return p.timeout
	}
	return time.Second * 60
}

func (p *Peer) GetGrpcUrl() string {
	if p.target == "" {
		p.target = utils.ToAddress(p.url)
	}
	return p.target
}

func (p *Peer) GetGrpcOpts() []grpc.DialOption {
	// Construct dialer options for the connection
	if p.grpcOpts != nil && len(p.grpcOpts) > 0 {
		return p.grpcOpts
	}
	if p.keepaliveParams == nil {
		p.keepaliveParams = GetDefaultKeepaliveParams()
	}
	var grpcOpts []grpc.DialOption
	if p.keepaliveParams.Time > 0 {
		grpcOpts = append(grpcOpts, grpc.WithKeepaliveParams(*p.keepaliveParams))
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.WaitForReady(!p.failFast)))
	if utils.AttemptSecured(p.url, p.inSecure) {
		tlsConfig := &tls.Config{
			RootCAs:      p.tlsCaCerts,
			Certificates: p.tlsClientCerts,
			ServerName:   p.serverName,
		}
		//verify if caCertificate was expired or not yet valid
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return certs2.VerifyCertificate(rawCerts, verifiedChains)
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize), grpc.MaxCallSendMsgSize(maxCallSendMsgSize)))
	p.grpcOpts = grpcOpts
	return p.grpcOpts
}

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *Peer) ProcessTransactionProposal(ctx context.Context, request *peer.SignedProposal) (*TransactionProposalResponse, error) {
	logger.Debugf("Processing proposal using endorser: %s", p.GetGrpcUrl())

	resp, err := retry2.RetryableInvoke(p.retryOpts,
		func() (interface{}, error) {
			return p.sendProposal(ctx, request)
		},
	)

	if err != nil {
		tpr := TransactionProposalResponse{
			Endorser: p.GetGrpcUrl(),
		}
		return &tpr, errors.Wrapf(err, "Transaction processing for endorser [%s]", p.GetGrpcUrl())
	}

	proposalResponse := resp.(*peer.ProposalResponse)

	chaincodeStatus, err := getChaincodeResponseStatus(proposalResponse)
	if err != nil {
		return nil, errors.WithMessage(err, "chaincode response status parsing failed")
	}

	tpr := TransactionProposalResponse{
		ProposalResponse: proposalResponse,
		Endorser:         p.GetGrpcUrl(),
		ChaincodeStatus:  chaincodeStatus,
		Status:           proposalResponse.Response.Status,
	}
	return &tpr, nil
}

func (p *Peer) sendProposal(ctx context.Context, proposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, p.GetTimeout())
	defer cancel()

	conn, err := grpcutils.DialContext(ctx, p.GetGrpcUrl(), p.GetGrpcOpts()...)
	if err != nil {
		return nil, ParseGrpcError(err, status2.EndorserClientStatus, p.GetGrpcUrl())
	}
	defer grpcutils.ReleaseConn(conn)

	endorserClient := peer.NewEndorserClient(conn)

	resp, err := endorserClient.ProcessProposal(ctx, proposal)

	if err != nil {
		logger.Errorf("process proposal failed [%s]", err)
		err = ParseGrpcError(err, status2.EndorserClientStatus, p.GetGrpcUrl())
	} else {
		//check error from response (for :fabric v1.2 and later)
		err = extractChaincodeErrorFromResponse(resp)
	}

	return resp, err
}

//extractChaincodeErrorFromResponse extracts chaincode error from proposal response
func extractChaincodeErrorFromResponse(resp *peer.ProposalResponse) error {
	if resp.Response.Status < int32(common.Status_SUCCESS) || resp.Response.Status >= int32(common.Status_BAD_REQUEST) {
		if chaincodeNotFoundPattern.MatchString(resp.Response.Message) {
			return status2.New(status2.EndorserClientStatus, int32(status2.ChaincodeNameNotFound), resp.Response.Message)
		}
		return status2.New(status2.ChaincodeStatus, resp.Response.Status, resp.Response.Message)
	}
	return nil
}

// getChaincodeResponseStatus gets the actual response status from response.Payload.extension.Response.status, as fabric always returns actual 200
func getChaincodeResponseStatus(response *peer.ProposalResponse) (int32, error) {
	if response.Payload != nil {
		payload, err := UnmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return 0, errors.Wrap(err, "unmarshal of proposal response payload failed")
		}

		extension, err := UnmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return 0, errors.Wrap(err, "unmarshal of chaincode action failed")
		}

		if extension != nil && extension.Response != nil {
			return extension.Response.Status, nil
		}
	}
	return response.Response.Status, nil
}

func (p *Peer) SetServerName(serverName string) *Peer {
	p.serverName = serverName
	return p
}

func (p *Peer) SetMspID(mspID string) *Peer {
	p.mspID = mspID
	return p
}

func (p *Peer) SetUrl(url string) *Peer {
	p.url = url
	p.target = ""
	return p
}

func (p *Peer) SetKeepaliveParams(keepaliveParams *keepalive.ClientParameters) *Peer {
	p.keepaliveParams = keepaliveParams
	return p
}

func (p *Peer) SetKeepaliveParamsTimeout(timeout time.Duration) *Peer {
	p.keepaliveParams.Timeout = timeout
	return p
}

func (p *Peer) SetKeepaliveParamsTime(time time.Duration) *Peer {
	p.keepaliveParams.Time = time
	return p
}

func (p *Peer) SetKeepaliveParamsPermitWithoutStream(permitWithoutStream bool) *Peer {
	p.keepaliveParams.PermitWithoutStream = permitWithoutStream
	return p
}

func (p *Peer) SetFailFast(failFast bool) *Peer {
	p.failFast = failFast
	return p
}

func (p *Peer) SetInSecure(inSecure bool) *Peer {
	p.inSecure = inSecure
	return p
}

func (p *Peer) SetProperties(properties map[string]interface{}) *Peer {
	p.properties = properties
	return p
}

func (p *Peer) SetTimeout(timeout time.Duration) *Peer {
	p.timeout = timeout
	return p
}

func (p *Peer) SetTlsCaCerts(tlsCaCerts *x509.CertPool) *Peer {
	p.tlsCaCerts = tlsCaCerts
	return p
}

func (p *Peer) AddTlsCaCerts(cert *x509.Certificate) *Peer {
	if p.tlsCaCerts == nil {
		p.tlsCaCerts = x509.NewCertPool()
	}
	p.tlsCaCerts.AddCert(cert)
	return p
}

func (p *Peer) AddTlsCaCertsOfPem(caCert string) *Peer {
	cert, e := certs2.PemToPublicKey([]byte(caCert))
	if e == nil {
		p.AddTlsCaCerts(cert)
	}
	return p
}

func (p *Peer) SetTlsClientCerts(tlsClientCerts []tls.Certificate) *Peer {
	p.tlsClientCerts = tlsClientCerts
	return p
}

// UnmarshalProposalResponsePayload unmarshals bytes to a ProposalResponsePayload
func UnmarshalProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, errors.Wrap(err, "error unmarshaling ProposalResponsePayload")
}

// UnmarshalChaincodeAction unmarshals bytes to a ChaincodeAction
func UnmarshalChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, errors.Wrap(err, "error unmarshaling ChaincodeAction")
}
