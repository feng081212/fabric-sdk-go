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
	"fabric-sdk-go/fabric/bccsp"
	"fmt"
)

func NewAesKeyGenerator(length int) *AesKeyGenerator {
	return &AesKeyGenerator{length: length}
}

type AesKeyGenerator struct {
	length int
}

func (kg *AesKeyGenerator) KeyGen(opts bccsp.KeyGenOpts) (bccsp.Key, error) {
	lowLevelKey, err := GetRandomBytes(kg.length)
	if err != nil {
		return nil, fmt.Errorf("failed generating AES %d key [%s]", kg.length, err)
	}

	return &AesPrivateKey{lowLevelKey, false}, nil
}
