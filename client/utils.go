package client

import (
	"crypto/rand"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
)

const (
	// NonceSize is the default NonceSize
	NonceSize = 24
)

// GetRandomBytes returns len random looking bytes
func GetRandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)

	// TODO: rand could fill less bytes then len
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "error getting random bytes")
	}

	return key, nil
}

// GetRandomNonce returns a random byte array of length NonceSize
func GetRandomNonce() ([]byte, error) {
	return GetRandomBytes(NonceSize)
}

// GetConfigUpdateFromEnvelope extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope'.
func GetConfigUpdateFromEnvelope(configEnvelope []byte) ([]byte, error) {

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope payload failed")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config update envelope")
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}