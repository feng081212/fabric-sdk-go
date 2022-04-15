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
	"crypto/rand"
	"fabric-sdk-go/fabric/bccsp"
)

type EcdsaSigner struct{}

func (p *EcdsaSigner) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) ([]byte, error) {
	kk := k.(*EcdsaPrivateKey).PrivateKey

	r, s, err := ecdsa.Sign(rand.Reader, kk, digest)
	if err != nil {
		return nil, err
	}

	s, err = bccsp.ToLowS(&kk.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return bccsp.MarshalECDSASignature(r, s)
}
