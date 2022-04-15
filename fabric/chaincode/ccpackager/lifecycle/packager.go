/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lifecycle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fabric-sdk-go/fabric/bccsp/hasher"
	"fabric-sdk-go/fabric/chaincode/platforms"
	"fmt"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"regexp"
)

const (
	codePackageName     = "code.tar.gz"
	metadataPackageName = "metadata.json"
)

// Descriptor holds the package data
type Descriptor struct {
	Path  string
	Type  peer.ChaincodeSpec_Type
	Label string
	Value interface{}
}

// Validate validates the package descriptor
func (p *Descriptor) Validate() error {
	if p.Path == "" {
		if p.Type != peer.ChaincodeSpec_EXTERNAL {
			return errors.New("chaincode path must be specified")
		}
	}

	if p.Type == peer.ChaincodeSpec_UNDEFINED {
		return errors.New("chaincode language must be specified")
	}

	if p.Label == "" {
		return errors.New("package label must be specified")
	}

	if err := ValidateLabel(p.Label); err != nil {
		return err
	}

	return nil
}

// PackageMetadata holds the path and type for a chaincode package
type PackageMetadata struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// ComputePackageID returns the package ID from the given label and install package
func ComputePackageID(label string, pkgBytes []byte) string {
	h := hasher.HashBySha256(pkgBytes)
	return fmt.Sprintf("%s:%x", label, h)
}

// NewCCPackage creates a chaincode package
func NewCCPackage(desc *Descriptor) ([]byte, error) {
	err := desc.Validate()
	if err != nil {
		return nil, err
	}

	pkgTarGzBytes, err := getTarGzBytes(desc)
	if err != nil {
		return nil, err
	}

	return pkgTarGzBytes, nil
}

func getTarGzBytes(desc *Descriptor) ([]byte, error) {
	payload := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(payload)
	tw := tar.NewWriter(gw)
	defer func() { _ = gw.Close() }()
	defer func() { _ = tw.Close() }()

	typeName := desc.Type.String()

	platform := platforms.GetPlatform(typeName)
	if platform == nil {
		return nil, fmt.Errorf("unknown chaincodeType: %s", typeName)
	}

	normalizedPath, err := platform.NormalizePath(desc.Path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to normalize chaincode path")
	}

	metadata := &PackageMetadata{
		Path:  normalizedPath,
		Type:  typeName,
		Label: desc.Label,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode package metadata into JSON")
	}

	err = writePackage(tw, metadataPackageName, metadataBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package metadata to tar")
	}
	codeBytes, err := platform.GetDeploymentPayload(desc.Path, desc.Value)
	if err != nil {
		return nil, errors.WithMessage(err, "error getting chaincode bytes")
	}
	err = writePackage(tw, codePackageName, codeBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error writing package code bytes to tar")
	}

	_ = tw.Close()
	_ = gw.Close()

	return payload.Bytes(), nil
}

func writePackage(tw *tar.Writer, name string, payload []byte) error {
	err := tw.WriteHeader(
		&tar.Header{
			Name: name,
			Size: int64(len(payload)),
			Mode: 0100644,
		},
	)
	if err != nil {
		return err
	}

	_, err = tw.Write(payload)
	return err
}

// LabelRegexp is the regular expression controlling the allowed characters
// for the package label.
var LabelRegexp = regexp.MustCompile(`^[[:alnum:]][[:alnum:]_.+-]*$`)

// ValidateLabel return an error if the provided label contains any invalid
// characters, as determined by LabelRegexp.
func ValidateLabel(label string) error {
	if !LabelRegexp.MatchString(label) {
		return errors.Errorf("invalid label '%s'. Label must be non-empty, can only consist of alphanumerics, symbols from '.+-_', and can only begin with alphanumerics", label)
	}

	return nil
}
