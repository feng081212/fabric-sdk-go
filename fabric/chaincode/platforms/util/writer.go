/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package util

import (
	"archive/tar"
	"bufio"
	"fmt"
	flogging "github.com/feng081212/fabric-sdk-go/common/logger"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/ccmetadata"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var logger = flogging.NewLogger("chaincode.platform.util")

// WriteFolderToTarPackage writes source files to a tarball.
// This utility is used for node js chaincode packaging, but not golang chaincode.
// Golang chaincode has more sophisticated file packaging, as implemented in golang/platform.go.
func WriteFolderToTarPackage(tw *tar.Writer, srcPath string, excludeDirs []string, includeFileTypeMap map[string]bool, excludeFileTypeMap map[string]bool) error {
	rootDirectory := filepath.Clean(srcPath)

	logger.Debug("writing folder to package", "rootDirectory", rootDirectory)

	var success bool
	walkFn := func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Errorf("Visit %s failed: %s", localPath, err)
			return err
		}

		if info.Mode().IsDir() {
			for _, excluded := range append(excludeDirs, ".git") {
				if info.Name() == excluded {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(localPath)
		if _, ok := includeFileTypeMap[ext]; includeFileTypeMap != nil && !ok {
			return nil
		}
		if excludeFileTypeMap[ext] {
			return nil
		}

		relpath, err := filepath.Rel(rootDirectory, localPath)
		if err != nil {
			return err
		}
		packagepath := filepath.ToSlash(relpath)

		// if file is metadata, keep the /META-INF directory, e.g: META-INF/statedb/couchdb/indexes/indexOwner.json
		// otherwise file is source code, put it in /src dir, e.g: src/marbles_chaincode.js
		if strings.HasPrefix(localPath, filepath.Join(rootDirectory, "META-INF")) {
			// Hidden files are not supported as metadata, therefore ignore them.
			// User often doesn't know that hidden files are there, and may not be able to delete them, therefore warn user rather than error out.
			if strings.HasPrefix(info.Name(), ".") {
				logger.Warnf("Ignoring hidden file in metadata directory: %s", packagepath)
				return nil
			}

			fileBytes, err := ioutil.ReadFile(localPath)
			if err != nil {
				return err
			}

			// Validate metadata file for inclusion in tar
			// Validation is based on the fully qualified path of the file
			err = ccmetadata.ValidateMetadataFile(packagepath, fileBytes)
			if err != nil {
				return err
			}
		} else { // file is not metadata, include in src
			packagepath = path.Join("src", packagepath)
		}

		err = WriteFileToPackage(localPath, packagepath, tw)
		if err != nil {
			return fmt.Errorf("Error writing file to package: %s", err)
		}

		success = true
		return nil
	}

	if err := filepath.Walk(rootDirectory, walkFn); err != nil {
		logger.Infof("Error walking rootDirectory: %s", err)
		return err
	}

	if !success {
		return errors.Errorf("no source files found in '%s'", srcPath)
	}
	return nil
}

// WriteFileToPackage writes a file to a tar stream.
func WriteFileToPackage(localpath string, packagepath string, tw *tar.Writer) error {
	logger.Debug("Writing file to tarball:", packagepath)
	fd, err := os.Open(localpath)
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}
	defer fd.Close()

	fi, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}

	header, err := tar.FileInfoHeader(fi, localpath)
	if err != nil {
		return fmt.Errorf("failed calculating FileInfoHeader: %s", err)
	}

	// Take the variance out of the tar by using zero time and fixed uid/gid.
	var zeroTime time.Time
	header.AccessTime = zeroTime
	header.ModTime = zeroTime
	header.ChangeTime = zeroTime
	header.Name = packagepath
	header.Mode = 0100644
	header.Uid = 500
	header.Gid = 500
	header.Uname = ""
	header.Gname = ""

	err = tw.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("failed to write header for %s: %s", localpath, err)
	}

	is := bufio.NewReader(fd)
	_, err = io.Copy(tw, is)
	if err != nil {
		return fmt.Errorf("failed to write %s as %s: %s", localpath, packagepath, err)
	}

	return nil
}

func WriteBytesToPackage(bs []byte, name string, tw *tar.Writer) error {
	zeroTime := time.Now()
	e := tw.WriteHeader(&tar.Header{
		Name:       name,
		Size:       int64(len(bs)),
		Mode:       0100644,
		AccessTime: zeroTime,
		ModTime:    zeroTime,
		ChangeTime: zeroTime,
		Uid:        500,
		Gid:        500,
	})
	if e != nil {
		return e
	}
	_, e = tw.Write(bs)
	return e
}

func Close(closer io.Closer) {
	if closer == nil {
		return
	}
	e := closer.Close()
	if e != nil {
		fmt.Println(fmt.Sprintf("%v close failure: %v", closer, e))
	}
}
