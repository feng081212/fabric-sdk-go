/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package retry

import (
	status2 "github.com/feng081212/fabric-sdk-go/fabric/errors/status"
	"time"

	"github.com/feng081212/fabric-protos-go/common"
	pb "github.com/feng081212/fabric-protos-go/peer"
	grpcCodes "google.golang.org/grpc/codes"
)

const (
	// DefaultAttempts number of retry attempts made by default
	DefaultAttempts = 3
	// DefaultInitialBackoff default initial backoff
	DefaultInitialBackoff = 500 * time.Millisecond
	// DefaultMaxBackoff default maximum backoff
	DefaultMaxBackoff = 60 * time.Second
	// DefaultBackoffFactor default backoff factor
	DefaultBackoffFactor = 2.0
)

// Resource Management Suggested Defaults
const (
	// ResMgmtDefaultAttempts number of retry attempts made by default
	ResMgmtDefaultAttempts = 5
	// ResMgmtDefaultInitialBackoff default initial backoff
	ResMgmtDefaultInitialBackoff = time.Second
	// ResMgmtDefaultMaxBackoff default maximum backoff
	ResMgmtDefaultMaxBackoff = 60 * time.Second
	// ResMgmtDefaultBackoffFactor default backoff factor
	ResMgmtDefaultBackoffFactor = 2.5
)

// DefaultOpts default retry options
var DefaultOpts = Opts{
	Attempts:       DefaultAttempts,
	InitialBackoff: DefaultInitialBackoff,
	MaxBackoff:     DefaultMaxBackoff,
	BackoffFactor:  DefaultBackoffFactor,
	RetryableCodes: DefaultRetryableCodes,
}

// DefaultChannelOpts default retry options for the channel client
var DefaultChannelOpts = Opts{
	Attempts:       DefaultAttempts,
	InitialBackoff: DefaultInitialBackoff,
	MaxBackoff:     DefaultMaxBackoff,
	BackoffFactor:  DefaultBackoffFactor,
	RetryableCodes: ChannelClientRetryableCodes,
}

// DefaultResMgmtOpts default retry options for the resource management client
var DefaultResMgmtOpts = Opts{
	Attempts:       ResMgmtDefaultAttempts,
	InitialBackoff: ResMgmtDefaultInitialBackoff,
	MaxBackoff:     ResMgmtDefaultMaxBackoff,
	BackoffFactor:  ResMgmtDefaultBackoffFactor,
	RetryableCodes: ResMgmtDefaultRetryableCodes,
}

// DefaultRetryableCodes these are the error codes, grouped by source of error,
// that are considered to be transient error conditions by default
var DefaultRetryableCodes = map[status2.Group][]status2.Code{
	status2.EndorserClientStatus: {
		status2.EndorsementMismatch,
		status2.ChaincodeNameNotFound,
	},
	status2.EndorserServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status2.OrdererServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status2.EventServerStatus: {
		status2.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status2.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status2.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status2.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status2.GRPCTransportStatus: {
		status2.Code(grpcCodes.Unavailable),
	},
}

// ResMgmtDefaultRetryableCodes are the suggested codes that should be treated as
// transient by github.com/feng081212/fabric-sdk-go/pkg/client/resmgmt.Client
var ResMgmtDefaultRetryableCodes = map[status2.Group][]status2.Code{
	status2.EndorserClientStatus: {
		status2.ConnectionFailed,
		status2.EndorsementMismatch,
		status2.ChaincodeNameNotFound,
	},
	status2.EndorserServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status2.OrdererServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
		status2.Code(common.Status_BAD_REQUEST),
		status2.Code(common.Status_NOT_FOUND),
	},
	status2.EventServerStatus: {
		status2.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status2.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status2.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status2.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status2.GRPCTransportStatus: {
		status2.Code(grpcCodes.Unavailable),
	},
}

// ChannelClientRetryableCodes are the suggested codes that should be treated as
// transient by github.com/feng081212/fabric-sdk-go/pkg/client/channel.Client
var ChannelClientRetryableCodes = map[status2.Group][]status2.Code{
	status2.EndorserClientStatus: {
		status2.ConnectionFailed, status2.EndorsementMismatch,
		status2.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status2.ChaincodeNameNotFound,
	},
	status2.EndorserServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
		status2.PvtDataDisseminationFailed,
	},
	status2.OrdererClientStatus: {
		status2.ConnectionFailed,
	},
	status2.OrdererServerStatus: {
		status2.Code(common.Status_SERVICE_UNAVAILABLE),
		status2.Code(common.Status_INTERNAL_SERVER_ERROR),
	},
	status2.EventServerStatus: {
		status2.Code(pb.TxValidationCode_DUPLICATE_TXID),
		status2.Code(pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE),
		status2.Code(pb.TxValidationCode_MVCC_READ_CONFLICT),
		status2.Code(pb.TxValidationCode_PHANTOM_READ_CONFLICT),
	},
	// TODO: gRPC introduced retries in v1.8.0. This can be replaced with the
	// gRPC fail fast option, once available
	status2.GRPCTransportStatus: {
		status2.Code(grpcCodes.Unavailable),
	},
}

// ChannelConfigRetryableCodes error codes to be taken into account for query channel config retry
var ChannelConfigRetryableCodes = map[status2.Group][]status2.Code{
	status2.EndorserClientStatus: {status2.EndorsementMismatch},
}
