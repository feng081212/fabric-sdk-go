/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package sw

import (
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/ecdsas"
	"reflect"
)

type X509PublicKeyImportOptsKeyImporter struct {
	bccsp *CSP
}

func (ki *X509PublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	x509Cert, ok := raw.(*x509.Certificate)
	if !ok {
		return nil, errors.New("invalid raw material. Expected *x509.Certificate")
	}

	pk := x509Cert.PublicKey

	switch pk.(type) {
	case *ecdsa.PublicKey:
		return ki.bccsp.KeyImporters[reflect.TypeOf(&ecdsas.ECDSAGoPublicKeyImportOpts{})].KeyImport(
			pk,
			&ecdsas.ECDSAGoPublicKeyImportOpts{Temporary: opts.Ephemeral()})
	default:
		return nil, errors.New("certificate's public key type not recognized. Supported keys: [ECDSA]")
	}
}
