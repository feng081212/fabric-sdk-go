package client

import (
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/hasher"
	"github.com/golang/protobuf/proto"
	"github.com/feng081212/fabric-protos-go/msp"
	"github.com/pkg/errors"
	"hash"
)

// User is a representation of a Fabric user
type User struct {
	ID            string
	MspID         string
	Certificate   []byte
	Key           bccsp.Key
	Bccsp         bccsp.BCCSP
	HashAlgorithm hasher.HashAlgorithm
}

// Verify a signature over some message using this identity as reference
func (u *User) Verify(digest []byte, signature []byte) error {
	b, e := u.Bccsp.Verify(u.Key, signature, digest, nil)
	if e != nil {
		return e
	}
	if !b {
		return errors.New("Verify failure")
	}
	return nil
}

// Serialize converts an identity to bytes
func (u *User) Serialize() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   u.MspID,
		IdBytes: u.Certificate,
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *User) EnrollmentCertificate() []byte {
	return u.Certificate
}

// Sign the message
func (u *User) Sign(msg []byte) ([]byte, error) {
	b, e := u.Hash(msg)
	if e != nil {
		return nil, e
	}
	return u.Bccsp.Sign(u.Key, b, nil)
}

func (u *User) GetHash() (h hash.Hash, err error) {
	return u.Bccsp.GetHash(u.HashAlgorithm)
}

func (u *User) Hash(msg []byte) (hash []byte, err error) {
	return u.Bccsp.Hash(msg, u.HashAlgorithm)
}
