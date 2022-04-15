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
package sw

import (
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"
	"fabric-sdk-go/fabric/bccsp"
	"fabric-sdk-go/fabric/bccsp/aess"
	"fabric-sdk-go/fabric/bccsp/ecdsas"
	"fabric-sdk-go/fabric/bccsp/hasher"
	"golang.org/x/crypto/sha3"
	"hash"
	"reflect"

	"github.com/pkg/errors"
)

// CSP provides a generic implementation of the BCCSP interface based
// on wrappers. It can be customized by providing implementations for the
// following algorithm-based wrappers: KeyGenerator, KeyDeriver, KeyImporter,
// Encryptor, Decryptor, Signer, Verifier, Hasher. Each wrapper is bound to a
// goland type representing either an option or a key.
type CSP struct {
	ks            bccsp.KeyStore
	hasher        hasher.Hasher
	KeyGenerators map[reflect.Type]KeyGenerator
	KeyDerivers   map[reflect.Type]KeyDeriver
	KeyImporters  map[reflect.Type]KeyImporter
	Encryptors    map[reflect.Type]Encryptor
	Decryptors    map[reflect.Type]Decryptor
	Signers       map[reflect.Type]Signer
	Verifiers     map[reflect.Type]Verifier
	Hashers       map[hasher.HashAlgorithm]hasher.Hasher
}

func NewCSP(keyStore bccsp.KeyStore, securityLevel int, hashFamily string) (*CSP, error) {
	if keyStore == nil {
		return nil, errors.Errorf("Invalid bccsp.KeyStore instance. It must be different from nil.")
	}

	// Init config
	conf := &config{}
	err := conf.setSecurityLevel(securityLevel, hashFamily)
	if err != nil {
		return nil, errors.Wrapf(err, "failed initializing configuration at [%v, %v]", securityLevel, hashFamily)
	}

	csp := &CSP{
		ks:            keyStore,
		hasher:        hasher.NewHasher(conf.hashFunction),
		KeyGenerators: make(map[reflect.Type]KeyGenerator),
		KeyDerivers:   make(map[reflect.Type]KeyDeriver),
		KeyImporters:  make(map[reflect.Type]KeyImporter),
		Encryptors:    make(map[reflect.Type]Encryptor),
		Decryptors:    make(map[reflect.Type]Decryptor),
		Signers:       make(map[reflect.Type]Signer),
		Verifiers:     make(map[reflect.Type]Verifier),
		Hashers:       make(map[hasher.HashAlgorithm]hasher.Hasher),
	}

	// Notice that errors are ignored here because some test will fail if one
	// of the following call fails.

	// Set the Encryptors
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AesPrivateKey{}), &aess.Aescbcpkcs7Encryptor{})

	// Set the Decryptor
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AesPrivateKey{}), &aess.Aescbcpkcs7Decryptor{})

	// Set the Signers
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.EcdsaPrivateKey{}), &ecdsas.EcdsaSigner{})

	// Set the Verifiers
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.EcdsaPrivateKey{}), &ecdsas.EcdsaPrivateKeyVerifier{})
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.EcdsaPublicKey{}), &ecdsas.EcdsaPublicKeyKeyVerifier{})

	// Set the Hasher
	csp.Hashers[hasher.SHA256] = hasher.NewHasher(sha256.New)
	csp.Hashers[hasher.SHA384] = hasher.NewHasher(sha512.New384)
	csp.Hashers[hasher.SHA3_256] = hasher.NewHasher(sha3.New256)
	csp.Hashers[hasher.SHA3_384] = hasher.NewHasher(sha3.New384)

	//_ = csp.AddWrapper(reflect.TypeOf(&SHAOpts{}), &bccsp.NewHasher({hash: )Conf.hashFunction})
	//_ = csp.AddWrapper(reflect.TypeOf(&SHA256Opts{}), &bccsp.hasher{hash: sha256.New})
	//_ = csp.AddWrapper(reflect.TypeOf(&SHA384Opts{}), &bccsp.hasher{hash: sha512.New384})
	//_ = csp.AddWrapper(reflect.TypeOf(&SHA3_256Opts{}), &bccsp.hasher{hash: sha3.New256})
	//_ = csp.AddWrapper(reflect.TypeOf(&SHA3_384Opts{}), &bccsp.hasher{hash: sha3.New384})

	// Set the key generators
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAKeyGenOpts{}), ecdsas.NewEcdsaKeyGenerator(conf.ellipticCurve))
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AESKeyGenOpts{}), aess.NewAesKeyGenerator(conf.aesBitLength))
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAP256KeyGenOpts{}), ecdsas.NewEcdsaKeyGenerator(elliptic.P256()))
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAP384KeyGenOpts{}), ecdsas.NewEcdsaKeyGenerator(elliptic.P384()))
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AES256KeyGenOpts{}), aess.NewAesKeyGenerator(32))
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AES192KeyGenOpts{}), aess.NewAesKeyGenerator(24))
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AES128KeyGenOpts{}), aess.NewAesKeyGenerator(16))

	// Set the key deriver
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.EcdsaPrivateKey{}), &ecdsas.EcdsaPrivateKeyKeyDeriver{})
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.EcdsaPublicKey{}), &ecdsas.EcdsaPublicKeyKeyDeriver{})
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AesPrivateKey{}), &aess.AesPrivateKeyKeyDeriver{HashFunction: conf.hashFunction, BitLength: conf.aesBitLength})

	// Set the key importers
	_ = csp.AddWrapper(reflect.TypeOf(&aess.AES256ImportKeyOpts{}), &aess.Aes256ImportKeyOptsKeyImporter{})
	_ = csp.AddWrapper(reflect.TypeOf(&aess.HMACImportKeyOpts{}), &aess.HmacImportKeyOptsKeyImporter{})
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAPKIXPublicKeyImportOpts{}), &ecdsas.EcdsaPKIXPublicKeyImportOptsKeyImporter{})
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAPrivateKeyImportOpts{}), &ecdsas.EcdsaPrivateKeyImportOptsKeyImporter{})
	_ = csp.AddWrapper(reflect.TypeOf(&ecdsas.ECDSAGoPublicKeyImportOpts{}), &ecdsas.EcdsaGoPublicKeyImportOptsKeyImporter{})
	_ = csp.AddWrapper(reflect.TypeOf(&X509PublicKeyImportOpts{}), &X509PublicKeyImportOptsKeyImporter{bccsp: csp})

	return csp, nil
}

// KeyGen generates a key using opts.
func (csp *CSP) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.New("Invalid Opts parameter. It must not be nil.")
	}

	keyGenerator, found := csp.KeyGenerators[reflect.TypeOf(opts)]
	if !found {
		return nil, errors.Errorf("Unsupported 'KeyGenOpts' provided [%v]", opts)
	}

	k, err = keyGenerator.KeyGen(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed generating key with opts [%v]", opts)
	}

	// If the key is not Ephemeral, store it.
	if !opts.Ephemeral() {
		// Store the key
		err = csp.ks.StoreKey(k)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed storing key [%s]", opts.Algorithm())
		}
	}

	return k, nil
}

// KeyDeriv derives a key from k using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *CSP) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}
	if opts == nil {
		return nil, errors.New("Invalid opts. It must not be nil.")
	}

	keyDeriver, found := csp.KeyDerivers[reflect.TypeOf(k)]
	if !found {
		return nil, errors.Errorf("Unsupported 'Key' provided [%v]", k)
	}

	k, err = keyDeriver.KeyDeriv(k, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed deriving key with opts [%v]", opts)
	}

	// If the key is not Ephemeral, store it.
	if !opts.Ephemeral() {
		// Store the key
		err = csp.ks.StoreKey(k)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed storing key [%s]", opts.Algorithm())
		}
	}

	return k, nil
}

// KeyImport imports a key from its raw representation using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *CSP) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if raw == nil {
		return nil, errors.New("Invalid raw. It must not be nil.")
	}
	if opts == nil {
		return nil, errors.New("Invalid opts. It must not be nil.")
	}

	keyImporter, found := csp.KeyImporters[reflect.TypeOf(opts)]
	if !found {
		return nil, errors.Errorf("Unsupported 'KeyImportOpts' provided [%v]", opts)
	}

	k, err = keyImporter.KeyImport(raw, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed importing key with opts [%v]", opts)
	}

	// If the key is not Ephemeral, store it.
	if !opts.Ephemeral() {
		// Store the key
		err = csp.ks.StoreKey(k)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed storing imported key with opts [%v]", opts)
		}
	}

	return
}

// GetKey returns the key this CSP associates to
// the Subject Key Identifier ski.
func (csp *CSP) GetKey(ski []byte) (k bccsp.Key, err error) {
	k, err = csp.ks.GetKey(ski)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed getting key for SKI [%v]", ski)
	}

	return
}

// Hash hashes messages msg using options opts.
func (csp *CSP) Hash(msg []byte, hashAlgorithm hasher.HashAlgorithm) (digest []byte, err error) {
	// Validate arguments
	hh, e := csp.GetHasher(hashAlgorithm)
	if e != nil {
		return nil, e
	}

	digest = hh.Hash(msg)

	return
}

func (csp *CSP) GetHasher(hashAlgorithm hasher.HashAlgorithm) (hasher.Hasher, error) {
	// Validate arguments
	if hashAlgorithm == "" && csp.hasher == nil {
		return nil, errors.New("bccsp is not set default hash algorithm. in this case params hashAlgorithm must not be nil")
	}

	if hashAlgorithm != "" {
		h, found := csp.Hashers[hashAlgorithm]
		if !found {
			return nil, errors.Errorf("Unsupported Hash Algorithm provided [%v]", hashAlgorithm)
		}
		return h, nil
	}

	return csp.hasher, nil
}

// GetHash returns and instance of hash.Hash using options opts.
// If opts is nil then the default hash function is returned.
func (csp *CSP) GetHash(hashAlgorithm hasher.HashAlgorithm) (h hash.Hash, err error) {

	hh, e := csp.GetHasher(hashAlgorithm)
	if e != nil {
		return nil, e
	}

	return hh.GetHash(), nil
}

// Sign signs digest using key k.
// The opts argument should be appropriate for the primitive used.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest).
func (csp *CSP) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}
	if len(digest) == 0 {
		return nil, errors.New("Invalid digest. Cannot be empty.")
	}

	keyType := reflect.TypeOf(k)
	signer, found := csp.Signers[keyType]
	if !found {
		return nil, errors.Errorf("Unsupported 'SignKey' provided [%s]", keyType)
	}

	signature, err = signer.Sign(k, digest, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed signing with opts [%v]", opts)
	}

	return
}

// Verify verifies signature against key k and digest
func (csp *CSP) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	// Validate arguments
	if k == nil {
		return false, errors.New("Invalid Key. It must not be nil.")
	}
	if len(signature) == 0 {
		return false, errors.New("Invalid signature. Cannot be empty.")
	}
	if len(digest) == 0 {
		return false, errors.New("Invalid digest. Cannot be empty.")
	}

	verifier, found := csp.Verifiers[reflect.TypeOf(k)]
	if !found {
		return false, errors.Errorf("Unsupported 'VerifyKey' provided [%v]", k)
	}

	valid, err = verifier.Verify(k, signature, digest, opts)
	if err != nil {
		return false, errors.Wrapf(err, "Failed verifing with opts [%v]", opts)
	}

	return
}

// Encrypt encrypts plaintext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *CSP) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) ([]byte, error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}

	encryptor, found := csp.Encryptors[reflect.TypeOf(k)]
	if !found {
		return nil, errors.Errorf("Unsupported 'EncryptKey' provided [%v]", k)
	}

	return encryptor.Encrypt(k, plaintext, opts)
}

// Decrypt decrypts ciphertext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *CSP) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}

	decryptor, found := csp.Decryptors[reflect.TypeOf(k)]
	if !found {
		return nil, errors.Errorf("Unsupported 'DecryptKey' provided [%v]", k)
	}

	plaintext, err = decryptor.Decrypt(k, ciphertext, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed decrypting with opts [%v]", opts)
	}

	return
}

// AddWrapper binds the passed type to the passed wrapper.
// Notice that that wrapper must be an instance of one of the following interfaces:
// KeyGenerator, KeyDeriver, KeyImporter, Encryptor, Decryptor, Signer, Verifier, Hasher.
func (csp *CSP) AddWrapper(t reflect.Type, w interface{}) error {
	if t == nil {
		return errors.Errorf("type cannot be nil")
	}
	if w == nil {
		return errors.Errorf("wrapper cannot be nil")
	}
	switch dt := w.(type) {
	case KeyGenerator:
		csp.KeyGenerators[t] = dt
	case KeyImporter:
		csp.KeyImporters[t] = dt
	case KeyDeriver:
		csp.KeyDerivers[t] = dt
	case Encryptor:
		csp.Encryptors[t] = dt
	case Decryptor:
		csp.Decryptors[t] = dt
	case Signer:
		csp.Signers[t] = dt
	case Verifier:
		csp.Verifiers[t] = dt
	//case bccsp.Hasher:
	//	csp.Hashers[t] = dt
	default:
		return errors.Errorf("wrapper type not valid, must be on of: KeyGenerator, KeyDeriver, KeyImporter, Encryptor, Decryptor, Signer, Verifier, Hasher")
	}
	return nil
}
