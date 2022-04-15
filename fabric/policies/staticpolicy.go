package policies

import (
	"github.com/golang/protobuf/proto"
	"github.com/feng081212/fabric-protos-go/common"
	"github.com/feng081212/fabric-protos-go/msp"
)

// AcceptAllPolicy always evaluates to true
var AcceptAllPolicy *common.SignaturePolicyEnvelope

// MarshaledAcceptAllPolicy is the Marshaled version of AcceptAllPolicy
var MarshaledAcceptAllPolicy []byte

// RejectAllPolicy always evaluates to false
var RejectAllPolicy *common.SignaturePolicyEnvelope

// MarshaledRejectAllPolicy is the Marshaled version of RejectAllPolicy
var MarshaledRejectAllPolicy []byte

func init() {
	AcceptAllPolicy = Envelope(NOutOf(0, []*common.SignaturePolicy{}), [][]byte{})
	MarshaledAcceptAllPolicy = protoMarshalOrPanic(AcceptAllPolicy)

	RejectAllPolicy = Envelope(NOutOf(1, []*common.SignaturePolicy{}), [][]byte{})
	MarshaledRejectAllPolicy = protoMarshalOrPanic(RejectAllPolicy)
}

// Envelope builds an envelope message embedding a SignaturePolicy
func Envelope(policy *common.SignaturePolicy, identities [][]byte) *common.SignaturePolicyEnvelope {
	ids := make([]*msp.MSPPrincipal, len(identities))
	for i := range ids {
		ids[i] = &msp.MSPPrincipal{PrincipalClassification: msp.MSPPrincipal_IDENTITY, Principal: identities[i]}
	}

	return &common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       policy,
		Identities: ids,
	}
}

// protoMarshalOrPanic serializes a protobuf message and panics if this
// operation fails
func protoMarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}

	return data
}
