package client

import (
	"github.com/feng081212/fabric-protos-go/peer"
	"hash"
)

// AnchorPeer encodes the necessary fields to identify an anchor peer.
type AnchorPeer struct {
	Host string
	Port int
}

// Policy encodes a channel config policy
type Policy struct {
	Type string
	Rule string
}

// Signer defines the interface needed for signing messages
type Signer interface {
	Sign(msg []byte) ([]byte, error)
	Serialize() ([]byte, error)
	Hash(msg []byte) (hash []byte, err error)
	GetHash() (h hash.Hash, err error)
}

// ApproveChaincodeRequest contains the parameters required to approve a chaincode
type ApproveChaincodeRequest struct {
	Name                string
	Version             string
	PackageID           string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     string
	ChannelConfigPolicy string
	CollectionConfig    []*peer.CollectionConfig
	InitRequired        bool
}

// QueryApprovedChaincodeRequest contains the parameters for an approved chaincode query
type QueryApprovedChaincodeRequest struct {
	Name     string
	Sequence int64
}

// CommitChaincodeRequest contains the parameters for a commit chaincode request
type CommitChaincodeRequest struct {
	Name                string
	Version             string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     string
	ChannelConfigPolicy string
	CollectionConfig    []*peer.CollectionConfig
	InitRequired        bool
}

// CheckChaincodeCommitReadinessRequest contains the parameters for checking the 'commit readiness' of a chaincode
type CheckChaincodeCommitReadinessRequest struct {
	Name                string
	Version             string
	Sequence            int64
	EndorsementPlugin   string
	ValidationPlugin    string
	SignaturePolicy     string
	ChannelConfigPolicy string
	CollectionConfig    []*peer.CollectionConfig
	InitRequired        bool
}

// QueryCommittedChaincodesRequest contains the parameters to query committed chaincodes.
// If name is not provided then all committed chaincodes on the given channel are returned,
// otherwise only the chaincode with the given name is returned.
type QueryCommittedChaincodesRequest struct {
	Name string
}
