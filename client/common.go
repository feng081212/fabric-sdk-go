package client

import (
	"crypto/tls"
	"encoding/hex"
	"fabric-sdk-go/fabric/bccsp/hasher"
	"fabric-sdk-go/fabric/endpoints"
	"fabric-sdk-go/fabric/policies"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"hash"
	"time"
)

// CreateChaincodeInvokeProposal creates a proposal for transaction.
func CreateChaincodeInvokeProposal(channelID string, signer Signer, request *endpoints.ChaincodeInvokeRequest) (*peer.Proposal, *common.Header, error) {
	if request.ChaincodeID == "" {
		return nil, nil, errors.New("ChaincodeID is required")
	}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "identity serialize failed")
	}

	h, err := signer.GetHash()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "hash function creation failed")
	}

	var args [][]byte
	// Add function name to arguments
	if request.Fcn != "" {
		args = append(args, []byte(request.Fcn))
		args = append(args, request.Args...)
	} else {
		args = request.Args
	}

	// create invocation spec to target a chaincode with arguments
	chaincodeInvocationSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type: request.Lang,
			ChaincodeId: &peer.ChaincodeID{
				Name: request.ChaincodeID,
			},
			Input: &peer.ChaincodeInput{
				Args:   args,
				IsInit: request.IsInit,
			},
		},
	}

	header, _ := CreateHeader(common.HeaderType_ENDORSER_TRANSACTION, channelID, h, creator, func(channelHeader *common.ChannelHeader) {
		channelHeader.Extension = ProtoMarshalIgnoreError(&peer.ChaincodeHeaderExtension{
			ChaincodeId: chaincodeInvocationSpec.ChaincodeSpec.ChaincodeId,
		})
	})

	payload := &peer.ChaincodeProposalPayload{
		Input:        ProtoMarshalIgnoreError(chaincodeInvocationSpec),
		TransientMap: request.TransientMap,
	}

	return &peer.Proposal{
		Header:  ProtoMarshalIgnoreError(header),
		Payload: ProtoMarshalIgnoreError(payload),
	}, header, nil
}

func CreatePayload(headerType common.HeaderType, channelID string, h hash.Hash, creator, data []byte, chHandler func(*common.ChannelHeader)) *common.Payload {

	header, _ := CreateHeader(headerType, channelID, h, creator, chHandler)

	payload := common.Payload{
		Header: header,
		Data:   data,
	}

	return &payload
}

func CreateHeader(headerType common.HeaderType, channelID string, h hash.Hash, creator []byte, chHandler func(*common.ChannelHeader)) (*common.Header, string) {
	nonce, _ := GetRandomNonce()
	txnID := ComputeTxnID(nonce, creator, h)
	channelHeader := &common.ChannelHeader{
		Type:      int32(headerType),
		ChannelId: channelID,
		TxId:      txnID,
		Epoch:     0,
		Timestamp: &timestamp.Timestamp{Seconds: time.Now().Unix(), Nanos: 0},
	}

	if chHandler != nil {
		chHandler(channelHeader)
	}

	return &common.Header{
		SignatureHeader: CreateSignatureHeader(creator, nonce),
		ChannelHeader:   ProtoMarshalIgnoreError(channelHeader),
	}, txnID
}

func CreateSignatureHeader(creator, nonce []byte) []byte {
	if nonce == nil {
		nonce, _ = GetRandomNonce()
	}
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	return ProtoMarshalIgnoreError(signatureHeader)
}

func ComputeTxnID(nonce, creator []byte, h hash.Hash) string {
	h.Write(nonce)
	h.Write(creator)
	return hex.EncodeToString(h.Sum(nil))
}

// signPayload signs payload
func signPayload(signer Signer, payload *common.Payload) (*endpoints.SignedEnvelope, error) {
	payloadBytes, signature, err := Sign(signer, payload)
	if err != nil {
		return nil, errors.WithMessage(err, "marshaling of payload failed")
	}
	return &endpoints.SignedEnvelope{Payload: payloadBytes, Signature: signature}, nil
}

func Sign(signer Signer, value proto.Message) ([]byte, []byte, error) {
	payloadBytes, err := proto.Marshal(value)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "marshaling of payload failed")
	}

	signature, err := signer.Sign(payloadBytes)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "signing of payload failed")
	}
	return payloadBytes, signature, nil
}

// tlsCertHash is a utility method to calculate the SHA256 hash of the configured certificate (for usage in channel headers)
func tlsCertHash(certs []tls.Certificate) ([]byte, error) {
	if len(certs) == 0 {
		return computeHash([]byte(""))
	}

	cert := certs[0]
	if len(cert.Certificate) == 0 {
		return computeHash([]byte(""))
	}

	return computeHash(cert.Certificate[0])
}

//computeHash computes hash for given bytes using underlying cryptosuite default
func computeHash(msg []byte) ([]byte, error) {
	h, err := GetDefaultBCCSP().Hash(msg, hasher.SHA256)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compute tls cert hash")
	}
	return h, err
}

func newSpecificSeekPosition(index uint64) *orderer.SeekPosition {
	return &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: index,
			},
		},
	}
}

func newNewestSeekPosition() *orderer.SeekPosition {
	return &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}
}

// GetLastConfigIndexFromBlock retrieves the index of the last config block as
// encoded in the block metadata
func GetLastConfigIndexFromBlock(block *common.Block) (uint64, error) {
	m, err := GetMetadataFromBlock(block, common.BlockMetadataIndex_SIGNATURES)
	if err != nil {
		return 0, errors.WithMessage(err, "failed to retrieve metadata")
	}
	// TODO FAB-15864 Remove this fallback when we can stop supporting upgrade from pre-1.4.1 orderer
	if len(m.Value) == 0 {
		m, err := GetMetadataFromBlock(block, common.BlockMetadataIndex_LAST_CONFIG)
		if err != nil {
			return 0, errors.WithMessage(err, "failed to retrieve metadata")
		}
		lc := &common.LastConfig{}
		err = proto.Unmarshal(m.Value, lc)
		if err != nil {
			return 0, errors.Wrap(err, "error unmarshalling LastConfig")
		}
		return lc.Index, nil
	}

	obm := &common.OrdererBlockMetadata{}
	err = proto.Unmarshal(m.Value, obm)
	if err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal orderer block metadata")
	}
	return obm.LastConfig.Index, nil
}

// GetMetadataFromBlock retrieves metadata at the specified index.
func GetMetadataFromBlock(block *common.Block, index common.BlockMetadataIndex) (*common.Metadata, error) {
	if block.Metadata == nil {
		return nil, errors.New("no metadata in block")
	}

	if len(block.Metadata.Metadata) <= int(index) {
		return nil, errors.Errorf("no metadata at index [%s]", index)
	}

	md := &common.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[index], md)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling metadata at index [%s]", index)
	}
	return md, nil
}

func ProtoMarshalIgnoreError(message proto.Message) (bytes []byte) {
	if message == nil {
		return nil
	}
	bytes, _ = proto.Marshal(message)
	return
}

// ConcatenateBytes is useful for combining multiple arrays of bytes, especially for
// signatures or digests over multiple fields
func ConcatenateBytes(data ...[]byte) []byte {
	finalLength := 0
	for _, slice := range data {
		finalLength += len(slice)
	}
	result := make([]byte, finalLength)
	last := 0
	for _, slice := range data {
		for i := range slice {
			result[i+last] = slice[i]
		}
		last += len(slice)
	}
	return result
}

func CapabilitiesConfigValue(capabilities []string) *common.Capabilities {
	cm := make(map[string]*common.Capability)
	for _, capability := range capabilities {
		cm[capability] = &common.Capability{}
	}
	return &common.Capabilities{Capabilities: cm}
}

func AddPolicies(cg *common.ConfigGroup, policyMap map[string]*Policy, modPolicy string) error {
	if policyMap == nil {
		return errors.Errorf("no policies defined")
	}
	if policyMap[AdminsPolicyKey] == nil {
		return errors.Errorf("no Admins policy defined")
	}
	if policyMap[ReadersPolicyKey] == nil {
		return errors.Errorf("no Readers policy defined")
	}
	if policyMap[WritersPolicyKey] == nil {
		return errors.Errorf("no Writers policy defined")
	}

	for policyName, policy := range policyMap {
		switch policy.Type {
		case ImplicitMetaPolicyType:
			pv, err := GetImplicitMetaPolicy(policy.Rule)
			if err != nil {
				return errors.Wrapf(err, "invalid implicit meta policy rule '%s'", policy.Rule)
			}
			addPolicy(cg, modPolicy, policyName, pv)
		case SignaturePolicyType:
			pv, err := GetSignaturePolicy(policy.Rule)
			if err != nil {
				return errors.Wrapf(err, "invalid signature policy rule '%s'", policy.Rule)
			}
			addPolicy(cg, modPolicy, policyName, pv)
		default:
			return errors.Errorf("unknown policy type: %s", policy.Type)
		}
	}
	return nil
}

func GetSignaturePolicy(rule string) (*common.Policy, error) {
	sp, err := policies.FromString(rule)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid signature policy rule '%s'", rule)
	}
	return NewSignaturePolicy(sp)
}

func GetImplicitMetaPolicy(rule string) (*common.Policy, error) {
	sp, err := policies.ImplicitMetaFromString(rule)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid implicit meta policy rule '%s'", rule)
	}
	return NewImplicitMetaPolicy(sp)
}

func NewImplicitMetaPolicy(policy *common.ImplicitMetaPolicy) (*common.Policy, error) {
	p, e := proto.Marshal(policy)
	if e != nil {
		return nil, e
	}
	return &common.Policy{
		Type:  int32(common.Policy_IMPLICIT_META),
		Value: p,
	}, nil
}

func NewSignaturePolicy(envelope *common.SignaturePolicyEnvelope) (*common.Policy, error) {
	acceptAllPolicy, e := proto.Marshal(envelope)
	if e != nil {
		return nil, e
	}
	return &common.Policy{
		Type:  int32(common.Policy_SIGNATURE),
		Value: acceptAllPolicy,
	}, nil
}

func addValue(cg *common.ConfigGroup, modPolicy string, key string, value proto.Message) error {
	val, err := proto.Marshal(value)
	if err != nil {
		return err
	}
	cg.Values[key] = &common.ConfigValue{
		Value:     val,
		ModPolicy: modPolicy,
	}
	return nil
}

func addPolicy(cg *common.ConfigGroup, modPolicy string, key string, policy *common.Policy) {
	cg.Policies[key] = &common.ConfigPolicy{Policy: policy, ModPolicy: modPolicy}
}

func CreatePolicyBytes(signaturePolicy, channelConfigPolicy string) (*peer.ApplicationPolicy, error) {

	if signaturePolicy != "" && channelConfigPolicy != "" {
		// mo policies, mo problems
		return nil, errors.New("cannot specify both \"--signature-policy\" and \"--channel-config-policy\"")
	}

	if signaturePolicy != "" {
		signaturePolicyEnvelope, err := policies.FromString(signaturePolicy)
		if err != nil {
			return nil, errors.Errorf("invalid signature policy: %s", signaturePolicy)
		}

		return &peer.ApplicationPolicy{
			Type: &peer.ApplicationPolicy_SignaturePolicy{
				SignaturePolicy: signaturePolicyEnvelope,
			},
		}, nil
	}

	if channelConfigPolicy != "" {
		return &peer.ApplicationPolicy{
			Type: &peer.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: channelConfigPolicy,
			},
		}, nil
	}

	return nil, nil
}
