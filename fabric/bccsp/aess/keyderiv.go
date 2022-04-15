/*
Copyright IBM Corp. 2017 All Rights Reserved.

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

package aess

import (
	"crypto/hmac"
	"errors"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"fmt"
	"hash"
)

type AesPrivateKeyKeyDeriver struct {
	HashFunction func() hash.Hash
	BitLength    int
}

func (kd *AesPrivateKeyKeyDeriver) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (bccsp.Key, error) {
	// Validate opts
	if opts == nil {
		return nil, errors.New("invalid opts parameter. It must not be nil")
	}

	aesK := k.(*AesPrivateKey)

	switch hmacOpts := opts.(type) {
	case *HMACTruncated256AESDeriveKeyOpts:
		mac := hmac.New(kd.HashFunction, aesK.PrivateKey)
		mac.Write(hmacOpts.Argument())
		return NewAesPrivateKey(mac.Sum(nil)[:kd.BitLength], false), nil

	case *HMACDeriveKeyOpts:
		mac := hmac.New(kd.HashFunction, aesK.PrivateKey)
		mac.Write(hmacOpts.Argument())
		return NewAesPrivateKey(mac.Sum(nil), true), nil

	default:
		return nil, fmt.Errorf("Unsupported 'KeyDerivOpts' provided [%v]", opts)
	}
}
