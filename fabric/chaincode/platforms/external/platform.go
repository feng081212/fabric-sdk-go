/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package external

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/feng081212/fabric-protos-go/peer"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/platforms/util"
	"github.com/pkg/errors"
	"os"
)

// Platform for chaincodes with external service
type Platform struct{}

// Name returns the name of this platform.
func (p *Platform) Name() string {
	return peer.ChaincodeSpec_EXTERNAL.String()
}

// ValidatePath is used to ensure that path provided points to something that
// looks like go chainccode.
//
// NOTE: this is only used at the _client_ side by the peer CLI.
func (p *Platform) ValidatePath(rawPath string) error {
	return nil
}

// NormalizePath is used to extract a relative module path from a module root.
//
// NOTE: this is only used at the _client_ side by the peer CLI.
func (p *Platform) NormalizePath(rawPath string) (string, error) {
	return "", nil
}

// ValidateCodePackage examines the chaincode archive to ensure it is valid.
//
// NOTE: this code is used in some transaction validation paths but can be changed
// post 2.0.
func (p *Platform) ValidateCodePackage(code []byte) error {
	return nil
}

// GetDeploymentPayload creates a gzip compressed tape archive that contains the
// required assets to build and run go chaincode.
//
// NOTE: this is only used at the _client_ side by the peer CLI.
func (p *Platform) GetDeploymentPayload(path string, val interface{}) ([]byte, error) {
	payload := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(payload)
	tw := tar.NewWriter(gw)

	defer util.Close(gw)
	defer util.Close(tw)

	var bs []byte
	var e error

	if val != nil {
		bs, e = json.Marshal(val)
		if e != nil {
			return nil, errors.Wrap(e, "failed to create tar for chaincode")
		}
	} else {
		fd, e := os.Open(path)
		if e != nil {
			return nil, errors.Wrap(e, "failed to create tar for chaincode")
		}
		defer util.Close(fd)
		buf := &bytes.Buffer{}
		_, e = buf.ReadFrom(fd)
		if e != nil {
			return nil, e
		}
		bs = buf.Bytes()
	}

	e = util.WriteBytesToPackage(bs, "connection.json", tw)
	if e != nil {
		return nil, e
	}

	_ = tw.Close()
	_ = gw.Close()
	return payload.Bytes(), nil
}
