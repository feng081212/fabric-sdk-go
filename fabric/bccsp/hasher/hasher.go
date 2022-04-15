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

package hasher

import (
	"crypto/sha256"
	"hash"
)

type HashAlgorithm string

const (
	SHA256   HashAlgorithm = "SHA256"
	SHA384   HashAlgorithm = "SHA384"
	SHA3_256 HashAlgorithm = "SHA3_256"
	SHA3_384 HashAlgorithm = "SHA3_384"
)

// Hasher is a BCCSP-like interface that provides hash algorithms
type Hasher interface {

	// Hash hashes messages msg using options opts.
	// If opts is nil, the default hash function will be used.
	Hash(msg []byte) (hash []byte)

	// GetHash returns and instance of hash.Hash using options opts.
	// If opts is nil, the default hash function will be returned.
	GetHash() (h hash.Hash)
}

func NewHasher(hashFunc func() hash.Hash) Hasher {
	return &hasher{hash: hashFunc}
}

type hasher struct {
	hash func() hash.Hash
}

func (c *hasher) Hash(msg []byte) []byte {
	return Hash(c.hash, msg)
}

func (c *hasher) GetHash() hash.Hash {
	return c.hash()
}

func Hash(hashFunc func() hash.Hash, msg []byte) []byte {
	h := hashFunc()
	h.Write(msg)
	return h.Sum(nil)
}

func HashBySha256(msg []byte) []byte  {
	return Hash(sha256.New, msg)
}
