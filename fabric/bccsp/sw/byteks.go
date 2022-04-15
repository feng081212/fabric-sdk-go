/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fabric-sdk-go/fabric/bccsp"
	"fabric-sdk-go/fabric/bccsp/ecdsas"
	"fmt"
)

// NewByteKeyStore instantiated  env key store at envs.
// The key store can be encrypted if a non-empty password is specified.
// It can be also be set as read only. In this case, any store operation
// will be forbidden
func NewByteKeyStore(pwd []byte, value []byte) (bccsp.KeyStore, error) {
	ks := &byteKeyStore{pwd: pwd, value: value}
	return ks, nil
}

// byteKeyStore is env key store.
type byteKeyStore struct {
	pwd   []byte
	value []byte
}

// ReadOnly returns true if this KeyStore is read only, false otherwise.
// If ReadOnly is true then StoreKey will fail.
func (ks *byteKeyStore) ReadOnly() bool {
	return true
}

// GetKey returns a key object whose SKI is the one passed.
func (ks *byteKeyStore) GetKey(ski []byte) (bccsp.Key, error) {
	key, err := pemToPrivateKey(ks.value, ks.pwd)
	if err != nil {
		return nil, err
	}
	var k bccsp.Key

	switch kk := key.(type) {
	case *ecdsa.PrivateKey:
		k = ecdsas.NewEcdsaPrivateKey(kk, true)
	}
	if k != nil && bytes.Equal(k.SKI(), ski) {
		return k, nil
	}
	return nil, fmt.Errorf("key with SKI %x do not match", ski)
}

// StoreKey stores the key k in this KeyStore.
// If this KeyStore is read only then the method will fail.
func (ks *byteKeyStore) StoreKey(k bccsp.Key) (err error) {
	return errors.New("read only KeyStore")
}
