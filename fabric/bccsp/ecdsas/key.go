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
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
)

func NewEcdsaPrivateKey(key *ecdsa.PrivateKey, exportable bool) *EcdsaPrivateKey {
	return &EcdsaPrivateKey{
		PrivateKey: key,
		exportable: exportable,
	}
}

func NewEcdsaPublicKey(key *ecdsa.PublicKey) *EcdsaPublicKey {
	return &EcdsaPublicKey{PubKey: key}
}

type EcdsaPrivateKey struct {
	PrivateKey *ecdsa.PrivateKey
	exportable bool
}

// Bytes converts this key to its byte representation,
// if this operation is allowed.
func (k *EcdsaPrivateKey) Bytes() ([]byte, error) {
	if !k.exportable {
		return nil, errors.New("not supported")
	}

	x509Encoded, err := x509.MarshalECPrivateKey(k.PrivateKey)
	if err != nil {
		return nil, err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return pemEncoded, nil
}

// SKI returns the subject key identifier of this key.
func (k *EcdsaPrivateKey) SKI() []byte {
	if k.PrivateKey == nil {
		return nil
	}

	// Marshall the public key
	raw := elliptic.Marshal(k.PrivateKey.Curve, k.PrivateKey.PublicKey.X, k.PrivateKey.PublicKey.Y)

	// Hash it
	hash := sha256.New()
	hash.Write(raw)
	return hash.Sum(nil)
}

// Symmetric returns true if this key is a symmetric key,
// false if this key is asymmetric
func (k *EcdsaPrivateKey) Symmetric() bool {
	return false
}

// Private returns true if this key is a private key,
// false otherwise.
func (k *EcdsaPrivateKey) Private() bool {
	return true
}

// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
// This method returns an error in symmetric key schemes.
func (k *EcdsaPrivateKey) PublicKey() (bccsp.Key, error) {
	return &EcdsaPublicKey{&k.PrivateKey.PublicKey}, nil
}

type EcdsaPublicKey struct {
	PubKey *ecdsa.PublicKey
}

// Bytes converts this key to its byte representation,
// if this operation is allowed.
func (k *EcdsaPublicKey) Bytes() (raw []byte, err error) {
	raw, err = x509.MarshalPKIXPublicKey(k.PubKey)
	if err != nil {
		return nil, fmt.Errorf("Failed marshalling key [%s]", err)
	}
	return
}

// SKI returns the subject key identifier of this key.
func (k *EcdsaPublicKey) SKI() []byte {
	if k.PubKey == nil {
		return nil
	}

	// Marshall the public key
	raw := elliptic.Marshal(k.PubKey.Curve, k.PubKey.X, k.PubKey.Y)

	// Hash it
	hash := sha256.New()
	hash.Write(raw)
	return hash.Sum(nil)
}

// Symmetric returns true if this key is a symmetric key,
// false if this key is asymmetric
func (k *EcdsaPublicKey) Symmetric() bool {
	return false
}

// Private returns true if this key is a private key,
// false otherwise.
func (k *EcdsaPublicKey) Private() bool {
	return false
}

// PublicKey returns the corresponding public key part of an asymmetric public/private key pair.
// This method returns an error in symmetric key schemes.
func (k *EcdsaPublicKey) PublicKey() (bccsp.Key, error) {
	return k, nil
}
