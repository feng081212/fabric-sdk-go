/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package platforms

import (
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/platforms/external"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/platforms/golang"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/platforms/java"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/platforms/node"
)

// Platform Interface for validating the specification and writing the package for
// the given platform
type Platform interface {
	Name() string
	ValidatePath(path string) error
	ValidateCodePackage(code []byte) error
	GetDeploymentPayload(path string, val interface{}) ([]byte, error)
	NormalizePath(path string) (string, error)
}

// SupportedPlatforms is the canonical list of platforms Fabric supports
var SupportedPlatforms = []Platform{
	&java.Platform{},
	&golang.Platform{},
	&node.Platform{},
	&external.Platform{},
}

func GetPlatform(chaincodeSpecType string) Platform {
	for _, platform := range SupportedPlatforms {
		if platform.Name() == chaincodeSpecType {
			return platform
		}
	}
	return nil
}
