package certs

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

func PemToPrivateKey(raw []byte) (interface{}, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM. Block must be different from nil [% x]", raw)
	}

	cert, err := DerToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, err
}

func DerToPrivateKey(der []byte) (key interface{}, err error) {

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

func PemToPublicKey(raw []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found in cert[% x]", raw)
	}

	return x509.ParseCertificate(block.Bytes)
}

func ReadPemFromString(pemString string) ([]byte, error) {
	if pemString == "" {
		return nil, errors.New("read pem error : pem value string can be nil")
	}
	bytes := []byte(pemString)
	b, _ := pem.Decode(bytes)
	if b == nil { // TODO: also check that the type is what we expect (cert vs key..)
		return nil, fmt.Errorf("read pem error : pem content is error [%s]", pemString)
	}
	return bytes, nil
}

func ReadPemMaterialFromStrings(values []string) ([][]byte, error) {
	if values == nil || len(values) == 0 {
		return nil, nil
	}
	var bytes [][]byte
	for _, v := range values {
		b, e := ReadPemFromString(v)
		if e != nil {
			return nil, e
		}
		bytes = append(bytes, b)
	}
	return bytes, nil
}
