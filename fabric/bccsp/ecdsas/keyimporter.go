package ecdsas

import (
	"crypto/ecdsa"
	"crypto/x509"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"fmt"
	"github.com/pkg/errors"
)

type EcdsaPKIXPublicKeyImportOptsKeyImporter struct{}

func (*EcdsaPKIXPublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("invalid raw material. Expected byte array")
	}

	if len(der) == 0 {
		return nil, errors.New("invalid raw. It must not be nil")
	}

	lowLevelKey, err := derToPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed converting PKIX to ECDSA public key [%s]", err)
	}

	ecdsaPK, ok := lowLevelKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed casting to ECDSA public key. Invalid raw material")
	}

	return NewEcdsaPublicKey(ecdsaPK), nil
}

type EcdsaPrivateKeyImportOptsKeyImporter struct{}

func (*EcdsaPrivateKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	der, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw material. Expected byte array")
	}

	if len(der) == 0 {
		return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw. It must not be nil")
	}

	lowLevelKey, err := derToPrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed converting PKIX to ECDSA public key [%s]", err)
	}

	ecdsaSK, ok := lowLevelKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed casting to ECDSA private key. Invalid raw material")
	}

	return NewEcdsaPrivateKey(ecdsaSK, true), nil
}

type EcdsaGoPublicKeyImportOptsKeyImporter struct{}

func (*EcdsaGoPublicKeyImportOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	lowLevelKey, ok := raw.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid raw material. Expected *ecdsa.PublicKey")
	}

	return NewEcdsaPublicKey(lowLevelKey), nil
}

func derToPublicKey(raw []byte) (pub interface{}, err error) {
	if len(raw) == 0 {
		return nil, errors.New("invalid DER. It must be different from nil")
	}

	key, err := x509.ParsePKIXPublicKey(raw)

	return key, err
}

func derToPrivateKey(der []byte) (key interface{}, err error) {

	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("invalid key type. The DER must contain an ecdsa.PrivateKey")
}
