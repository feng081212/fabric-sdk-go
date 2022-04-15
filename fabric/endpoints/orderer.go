/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoints

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fabric-sdk-go/common/utils"
	"fabric-sdk-go/common/utils/grpcutils"
	certs2 "fabric-sdk-go/fabric/crypto/certs"
	"fabric-sdk-go/fabric/errors/multi"
	retry2 "fabric-sdk-go/fabric/errors/retry"
	"fabric-sdk-go/fabric/errors/status"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"
	"io"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
)

// A SignedEnvelope can can be sent to an orderer for broadcasting
type SignedEnvelope struct {
	Payload   []byte
	Signature []byte
}

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	url             string
	target          string
	serverName      string
	grpcOpts        []grpc.DialOption
	keepaliveParams *keepalive.ClientParameters
	timeout         time.Duration
	failFast        bool
	allowInsecure   bool
	tlsCaCerts      *x509.CertPool
	tlsClientCerts []tls.Certificate
	retryOpts      retry2.Opts
}

func (o *Orderer) GetTlsClientCerts() []tls.Certificate {
	return o.tlsClientCerts
}

// URL Get the Orderer url. Required property for the instance objects.
// Returns the address of the Orderer.
func (o *Orderer) URL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) SendBroadcast(ctx context.Context, envelope *SignedEnvelope) (*common.Status, error) {
	res, err := retry2.RetryableInvoke(o.retryOpts,
		func() (interface{}, error) {
			return o.sendBroadcast(ctx, envelope)
		},
	)
	if err != nil {
		return nil, err
	}
	return res.(*common.Status), nil
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) sendBroadcast(ctx context.Context, envelope *SignedEnvelope) (*common.Status, error) {
	ctx, cancel := context.WithTimeout(ctx, o.GetTimeout())
	defer cancel()
	conn, err := grpcutils.DialContext(ctx, o.GetGrpcUrl(), o.GetGrpcOpts()...)
	if err != nil {
		return nil, ParseGrpcError(err, status.OrdererClientStatus, o.GetGrpcUrl())
	}
	defer grpcutils.ReleaseConn(conn)

	broadcastClient, err := ab.NewAtomicBroadcastClient(conn).Broadcast(ctx)
	if err != nil {
		return nil, ParseGrpcError(err, status.OrdererClientStatus, o.GetGrpcUrl())
	}

	responses := make(chan interface{})
	errs := make(chan error, 1)

	go broadcastStream(broadcastClient, responses, errs)

	err = broadcastClient.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to send envelope to orderer")
	}
	if err = broadcastClient.CloseSend(); err != nil {
		logger.Debugf("unable to close broadcast client [%s]", err)
	}

	res, err := wrapStreamResponseRPC(responses, errs)
	if err != nil {
		return nil, err
	}
	result := *(res.(*interface{}))
	sts := result.(common.Status)
	return &sts, nil
}

func broadcastStream(broadcastClient ab.AtomicBroadcast_BroadcastClient, responses chan interface{}, errs chan error) {
	for {
		broadcastResponse, err := broadcastClient.Recv()
		if err == io.EOF {
			// done
			close(responses)
			return
		}

		if err != nil {
			rpcStatus, ok := grpcstatus.FromError(err)
			if ok {
				err = status.NewFromGRPCStatus(rpcStatus)
			}
			//ParseGrpcError(err, )
			errs <- errors.Wrap(err, "broadcast recv failed")
			close(responses)
			return
		}

		if broadcastResponse.Status == common.Status_SUCCESS {
			responses <- broadcastResponse.Status
		} else {
			errs <- status.New(status.OrdererServerStatus, int32(broadcastResponse.Status), broadcastResponse.Info)
		}
	}
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) SendDeliver(ctx context.Context, envelope *SignedEnvelope) (*common.Block, error) {
	res, err := retry2.RetryableInvoke(o.retryOpts,
		func() (interface{}, error) {
			return o.sendDeliver(ctx, envelope)
		},
	)
	if err != nil {
		return nil, err
	}
	return res.(*common.Block), nil
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) sendDeliver(ctx context.Context, envelope *SignedEnvelope) (*common.Block, error) {

	ctx, cancel := context.WithTimeout(ctx, o.GetTimeout())
	defer cancel()

	conn, err := grpcutils.DialContext(ctx, o.GetGrpcUrl(), o.GetGrpcOpts()...)
	if err != nil {
		return nil, ParseGrpcError(err, status.OrdererClientStatus, o.GetGrpcUrl())
	}
	defer grpcutils.ReleaseConn(conn)

	// Create atomic broadcast client
	broadcastClient, err := ab.NewAtomicBroadcastClient(conn).Deliver(ctx)
	if err != nil {
		logger.Errorf("deliver failed [%s]", err)
		return nil, errors.Wrap(err, "deliver failed")
	}

	responses := make(chan interface{})
	errs := make(chan error, 1)

	// Receive blocks from the GRPC stream and put them on the channel
	go blockStream(broadcastClient, responses, errs)

	// Send block request envelope
	logger.Debug("Requesting blocks from ordering service")
	err = broadcastClient.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	})
	if err != nil {
		logger.Warnf("failed to send block request to orderer [%s]", err)
	}
	if err = broadcastClient.CloseSend(); err != nil {
		logger.Debugf("unable to close deliver client [%s]", err)
	}

	res, err := wrapStreamResponseRPC(responses, errs)
	if err != nil {
		return nil, err
	}
	r := *(res.(*interface{}))
	return r.(*common.Block), nil
}

func blockStream(deliverClient ab.AtomicBroadcast_DeliverClient, responses chan interface{}, errs chan error) {

	for {
		response, err := deliverClient.Recv()
		if err == io.EOF {
			// done
			close(responses)
			return
		}

		if err != nil {
			errs <- errors.Wrap(err, "recv from ordering service failed")
			close(responses)
			return
		}

		// Assert response type
		switch t := response.Type.(type) {
		// Seek operation success, no more responses
		case *ab.DeliverResponse_Status:
			logger.Debugf("Received deliver response status from ordering service: %s", t.Status)
			if t.Status != common.Status_SUCCESS {
				errs <- status.New(status.OrdererServerStatus, int32(t.Status), "error status from ordering service")
			}

		// Response is a requested block
		case *ab.DeliverResponse_Block:
			logger.Debug("Received block from ordering service")
			responses <- response.GetBlock()
		// Unknown response
		default:
			// ignore unknown types.
			logger.Infof("unknown response type from ordering service %T", t)
		}
	}
}

func wrapStreamResponseRPC(response chan interface{}, errs chan error) (interface{}, error) {
	// This function currently returns the last received block and error.
	var block interface{}
	var err multi.Errors

read:
	for {
		select {
		case b, ok := <-response:
			// We need to block until SendDeliver releases the connection. Currently
			// this is triggered by the go chan closing.
			// TODO: we may want to refactor (e.g., adding a synchronous SendDeliver)
			if !ok {
				break read
			}
			block = b
		case e := <-errs:
			err = append(err, e)
		}
	}

	// drain remaining errors.
	for i := 0; i < len(errs); i++ {
		e := <-errs
		err = append(err, e)
	}

	return &block, err.ToError()
}

func (p *Orderer) GetTimeout() time.Duration {
	if p.timeout != 0 {
		return p.timeout
	}
	return time.Second * 60
}

func (p *Orderer) GetGrpcUrl() string {
	if p.target == "" {
		p.target = utils.ToAddress(p.url)
	}
	return p.target
}

func (p *Orderer) GetGrpcOpts() []grpc.DialOption {
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
	if utils.AttemptSecured(p.url, p.allowInsecure) {
		//tls config
		tlsConfig := &tls.Config{
			RootCAs:      p.tlsCaCerts,
			Certificates: p.tlsClientCerts,
			ServerName:   p.serverName,
		}
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

func (p *Orderer) SetServerName(serverName string) *Orderer {
	p.serverName = serverName
	return p
}

func (p *Orderer) SetUrl(url string) *Orderer {
	p.url = url
	p.target = ""
	return p
}

func (p *Orderer) SetKeepaliveParams(keepaliveParams *keepalive.ClientParameters) *Orderer {
	p.keepaliveParams = keepaliveParams
	return p
}

func (p *Orderer) SetKeepaliveParamsTimeout(timeout time.Duration) *Orderer {
	p.keepaliveParams.Timeout = timeout
	return p
}

func (p *Orderer) SetKeepaliveParamsTime(time time.Duration) *Orderer {
	p.keepaliveParams.Time = time
	return p
}

func (p *Orderer) SetKeepaliveParamsPermitWithoutStream(permitWithoutStream bool) *Orderer {
	p.keepaliveParams.PermitWithoutStream = permitWithoutStream
	return p
}

func (p *Orderer) SetFailFast(failFast bool) *Orderer {
	p.failFast = failFast
	return p
}

func (p *Orderer) SetAllowInsecure(allowInsecure bool) *Orderer {
	p.allowInsecure = allowInsecure
	return p
}

func (p *Orderer) SetTimeout(timeout time.Duration) *Orderer {
	p.timeout = timeout
	return p
}

func (p *Orderer) SetTlsCaCerts(tlsCaCerts *x509.CertPool) *Orderer {
	p.tlsCaCerts = tlsCaCerts
	return p
}

func (p *Orderer) AddTlsCaCerts(cert *x509.Certificate) *Orderer {
	if p.tlsCaCerts == nil {
		p.tlsCaCerts = x509.NewCertPool()
	}
	p.tlsCaCerts.AddCert(cert)
	return p
}

func (p *Orderer) AddTlsCaCertsOfPem(caCert string) *Orderer {
	cert, e := certs2.PemToPublicKey([]byte(caCert))
	if e == nil {
		p.AddTlsCaCerts(cert)
	}
	return p
}

func (p *Orderer) SetTlsClientCerts(tlsClientCerts []tls.Certificate) *Orderer {
	p.tlsClientCerts = tlsClientCerts
	return p
}
