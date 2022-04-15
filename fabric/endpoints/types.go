package endpoints

import (
	"github.com/hyperledger/fabric-protos-go/peer"
)

// ChaincodeInvokeRequest contains the parameters for sending a transaction proposal.
// nolint: maligned
type ChaincodeInvokeRequest struct {
	ChaincodeID  string
	Lang         peer.ChaincodeSpec_Type
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
	IsInit       bool
}

// TransactionProposalResponse respresents the result of transaction proposal processing.
type TransactionProposalResponse struct {
	Endorser string
	// Status is the EndorserStatus
	Status int32
	// ChaincodeStatus is the status returned by Chaincode
	ChaincodeStatus int32
	*peer.ProposalResponse
}
