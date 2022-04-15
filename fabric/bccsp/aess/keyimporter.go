package aess

import (
	"fabric-sdk-go/fabric/bccsp"
	"fmt"
	"github.com/pkg/errors"
)

type Aes256ImportKeyOptsKeyImporter struct{}

func (*Aes256ImportKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	aesRaw, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("invalid raw material. Expected byte array")
	}

	if aesRaw == nil {
		return nil, errors.New("invalid raw material. It must not be nil")
	}

	if len(aesRaw) != 32 {
		return nil, fmt.Errorf("invalid Key Length [%d]. Must be 32 bytes", len(aesRaw))
	}

	return NewAesPrivateKey(aesRaw, false), nil
}

type HmacImportKeyOptsKeyImporter struct{}

func (*HmacImportKeyOptsKeyImporter) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (bccsp.Key, error) {
	aesRaw, ok := raw.([]byte)
	if !ok {
		return nil, errors.New("invalid raw material. Expected byte array")
	}

	if len(aesRaw) == 0 {
		return nil, errors.New("invalid raw material. It must not be nil")
	}

	return NewAesPrivateKey(aesRaw, false), nil
}
