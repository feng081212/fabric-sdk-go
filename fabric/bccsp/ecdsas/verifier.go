/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
package ecdsas

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
)

type EcdsaPrivateKeyVerifier struct{}

func (v *EcdsaPrivateKeyVerifier) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (bool, error) {
	return verifyECDSA(&(k.(*EcdsaPrivateKey).PrivateKey.PublicKey), signature, digest, opts)
}

type EcdsaPublicKeyKeyVerifier struct{}

func (v *EcdsaPublicKeyKeyVerifier) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (bool, error) {
	return verifyECDSA(k.(*EcdsaPublicKey).PubKey, signature, digest, opts)
}

func verifyECDSA(k *ecdsa.PublicKey, signature, digest []byte, opts bccsp.SignerOpts) (bool, error) {
	r, s, err := bccsp.UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("failed unmashalling signature [%s]", err)
	}

	lowS, err := bccsp.IsLowS(k, s)
	if err != nil {
		return false, err
	}

	if !lowS {
		return false, fmt.Errorf("Invalid S. Must be smaller than half the order [%s][%s].", s, bccsp.GetCurveHalfOrdersAt(k.Curve))
	}

	return ecdsa.Verify(k, digest, r, s), nil
}
