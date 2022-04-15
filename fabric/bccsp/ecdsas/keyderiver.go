package ecdsas

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"github.com/pkg/errors"
	"math/big"
)

type EcdsaPublicKeyKeyDeriver struct{}

func (kd *EcdsaPublicKeyKeyDeriver) KeyDeriv(key bccsp.Key, opts bccsp.KeyDerivOpts) (bccsp.Key, error) {
	// Validate opts
	if opts == nil {
		return nil, errors.New("Invalid opts parameter. It must not be nil.")
	}

	ecdsaK := key.(*EcdsaPublicKey)

	// Re-randomized an ECDSA private key
	reRandOpts, ok := opts.(*ECDSAReRandKeyOpts)
	if !ok {
		return nil, fmt.Errorf("Unsupported 'KeyDerivOpts' provided [%v]", opts)
	}

	tempSK := &ecdsa.PublicKey{
		Curve: ecdsaK.PubKey.Curve,
		X:     new(big.Int),
		Y:     new(big.Int),
	}

	var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
	var one = new(big.Int).SetInt64(1)
	n := new(big.Int).Sub(ecdsaK.PubKey.Params().N, one)
	k.Mod(k, n)
	k.Add(k, one)

	// Compute temporary public key
	tempX, tempY := ecdsaK.PubKey.ScalarBaseMult(k.Bytes())
	tempSK.X, tempSK.Y = tempSK.Add(
		ecdsaK.PubKey.X, ecdsaK.PubKey.Y,
		tempX, tempY,
	)

	// Verify temporary public key is a valid point on the reference curve
	isOn := tempSK.Curve.IsOnCurve(tempSK.X, tempSK.Y)
	if !isOn {
		return nil, errors.New("Failed temporary public key IsOnCurve check.")
	}

	return NewEcdsaPublicKey(tempSK), nil
}

type EcdsaPrivateKeyKeyDeriver struct{}

func (kd *EcdsaPrivateKeyKeyDeriver) KeyDeriv(key bccsp.Key, opts bccsp.KeyDerivOpts) (bccsp.Key, error) {
	// Validate opts
	if opts == nil {
		return nil, errors.New("Invalid opts parameter. It must not be nil.")
	}

	ecdsaK := key.(*EcdsaPrivateKey)

	// Re-randomized an ECDSA private key
	reRandOpts, ok := opts.(*ECDSAReRandKeyOpts)
	if !ok {
		return nil, fmt.Errorf("Unsupported 'KeyDerivOpts' provided [%v]", opts)
	}

	tempSK := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: ecdsaK.PrivateKey.Curve,
			X:     new(big.Int),
			Y:     new(big.Int),
		},
		D: new(big.Int),
	}

	var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
	var one = new(big.Int).SetInt64(1)
	n := new(big.Int).Sub(ecdsaK.PrivateKey.Params().N, one)
	k.Mod(k, n)
	k.Add(k, one)

	tempSK.D.Add(ecdsaK.PrivateKey.D, k)
	tempSK.D.Mod(tempSK.D, ecdsaK.PrivateKey.PublicKey.Params().N)

	// Compute temporary public key
	tempX, tempY := ecdsaK.PrivateKey.PublicKey.ScalarBaseMult(k.Bytes())
	tempSK.PublicKey.X, tempSK.PublicKey.Y =
		tempSK.PublicKey.Add(
			ecdsaK.PrivateKey.PublicKey.X, ecdsaK.PrivateKey.PublicKey.Y,
			tempX, tempY,
		)

	// Verify temporary public key is a valid point on the reference curve
	isOn := tempSK.Curve.IsOnCurve(tempSK.PublicKey.X, tempSK.PublicKey.Y)
	if !isOn {
		return nil, errors.New("failed temporary public key IsOnCurve check")
	}

	return NewEcdsaPrivateKey(tempSK, true), nil
}
