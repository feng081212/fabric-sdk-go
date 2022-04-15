package ecdsas

// ECDSA Elliptic Curve Digital Signature Algorithm (key gen, import, sign, verify),
// at default security level.
// Each BCCSP may or may not support default security level. If not supported than
// an error will be returned.
var ECDSA = "ECDSA"

// ECDSAReRand ECDSA key re-randomization
var ECDSAReRand = "ECDSA_RERAND"

// ECDSAP256 ECDSA Elliptic Curve Digital Signature Algorithm over P-256 curve
var ECDSAP256 = "ECDSAP256"

// ECDSAP384 ECDSA Elliptic Curve Digital Signature Algorithm over P-384 curve
var ECDSAP384 = "ECDSAP384"

// ECDSAReRandKeyOpts contains options for ECDSA key re-randomization.
type ECDSAReRandKeyOpts struct {
	Temporary bool
	Expansion []byte
}

// Algorithm returns the key derivation algorithm identifier (to be used).
func (opts *ECDSAReRandKeyOpts) Algorithm() string {
	return ECDSAReRand
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAReRandKeyOpts) Ephemeral() bool {
	return opts.Temporary
}

// ExpansionValue returns the re-randomization factor
func (opts *ECDSAReRandKeyOpts) ExpansionValue() []byte {
	return opts.Expansion
}


// ECDSAP256KeyGenOpts contains options for ECDSA key generation with curve P-256.
type ECDSAP256KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *ECDSAP256KeyGenOpts) Algorithm() string {
	return ECDSAP256
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAP256KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// ECDSAP384KeyGenOpts contains options for ECDSA key generation with curve P-384.
type ECDSAP384KeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *ECDSAP384KeyGenOpts) Algorithm() string {
	return ECDSAP384
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAP384KeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// ECDSAKeyGenOpts contains options for ECDSA key generation.
type ECDSAKeyGenOpts struct {
	Temporary bool
}

// Algorithm returns the key generation algorithm identifier (to be used).
func (opts *ECDSAKeyGenOpts) Algorithm() string {
	return ECDSA
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAKeyGenOpts) Ephemeral() bool {
	return opts.Temporary
}

// ECDSAPKIXPublicKeyImportOpts contains options for ECDSA public key importation in PKIX format
type ECDSAPKIXPublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *ECDSAPKIXPublicKeyImportOpts) Algorithm() string {
	return ECDSA
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAPKIXPublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

// ECDSAPrivateKeyImportOpts contains options for ECDSA secret key importation in DER format
// or PKCS#8 format.
type ECDSAPrivateKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *ECDSAPrivateKeyImportOpts) Algorithm() string {
	return ECDSA
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAPrivateKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}

// ECDSAGoPublicKeyImportOpts contains options for ECDSA key importation from ecdsa.PublicKey
type ECDSAGoPublicKeyImportOpts struct {
	Temporary bool
}

// Algorithm returns the key importation algorithm identifier (to be used).
func (opts *ECDSAGoPublicKeyImportOpts) Algorithm() string {
	return ECDSA
}

// Ephemeral returns true if the key to generate has to be ephemeral,
// false otherwise.
func (opts *ECDSAGoPublicKeyImportOpts) Ephemeral() bool {
	return opts.Temporary
}