package factory

import (
	"fabric-sdk-go/fabric/bccsp"
	"fabric-sdk-go/fabric/bccsp/sw"
	"github.com/pkg/errors"
)

// GetBccsp ...
func GetBccsp(bccspOpts *BccspOpts) (bccsp.BCCSP, error) {
	if bccspOpts == nil {
		return nil, errors.New("bccsp opts can not be nil")
	}
	switch bccspOpts.Default {
	case "SW", "sw", "Sw", "sW":
		return GetSwBccsp(bccspOpts.SW)
	default:
		return nil, errors.New("bccsp opts default supported list is [sw]")
	}
}

func GetSwBccsp(swOpts *SwOpts) (bccsp.BCCSP, error) {
	if swOpts == nil {
		return nil, errors.New("sw bccsp opts can not be nil")
	}
	var ks bccsp.KeyStore
	if swOpts.ByteKeyStore != nil && swOpts.ByteKeyStore.Value != "" {
		fks, err := sw.NewByteKeyStore([]byte(swOpts.ByteKeyStore.Password), []byte(swOpts.ByteKeyStore.Value))
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize software key store")
		}
		ks = fks
	} else if swOpts.FileKeystore != nil {
		fks, err := sw.NewFileBasedKeyStore(nil, swOpts.FileKeystore.KeyStorePath, false)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize software key store")
		}
		ks = fks
	} else if swOpts.InMemKeystore != nil {
		ks = sw.NewInMemoryKeyStore()
	} else {
		// Default to ephemeral key store
		ks = sw.NewDummyKeyStore()
	}

	return sw.NewCSP(ks, swOpts.SecLevel, swOpts.HashFamily)
}