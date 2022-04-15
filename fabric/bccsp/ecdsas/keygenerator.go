package ecdsas

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"fmt"
)

type EcdsaKeyGenerator struct {
	curve elliptic.Curve
}

func NewEcdsaKeyGenerator(curve elliptic.Curve) *EcdsaKeyGenerator {
	return &EcdsaKeyGenerator{curve: curve}
}

func (kg *EcdsaKeyGenerator) KeyGen(opts bccsp.KeyGenOpts) (bccsp.Key, error) {
	privKey, err := ecdsa.GenerateKey(kg.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", kg.curve, err)
	}
	return NewEcdsaPrivateKey(privKey, true), nil
}
