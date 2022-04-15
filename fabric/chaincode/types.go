package chaincode

import pb "github.com/feng081212/fabric-protos-go/peer"

// CCPackage contains package type and bytes required to create CDS
type CCPackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}
