package client

import (
	"crypto/ecdsa"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/ecdsas"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/factory"
	"github.com/feng081212/fabric-sdk-go/fabric/crypto/certs"
	"github.com/feng081212/fabric-sdk-go/fabric/endpoints"
	"github.com/hyperledger/fabric-protos-go/common"
)

func GetPeer(mspID, serviceName, url, caCertificate string) *endpoints.Peer {
	return endpoints.EmptyPeer().SetMspID(mspID).SetServerName(serviceName).SetUrl(url).AddTlsCaCertsOfPem(caCertificate)
}

func GetOrderer(serviceName, url, caCertificate string) *endpoints.Orderer {
	return endpoints.EmptyOrderer().SetServerName(serviceName).SetUrl(url).AddTlsCaCertsOfPem(caCertificate)
}

func GetUser(id, mspID, certificate, privateKey string) (*User, error) {
	key, e := certs.PemToPrivateKey([]byte(privateKey))
	if e != nil {
		return nil, e
	}
	k := ecdsas.NewEcdsaPrivateKey(key.(*ecdsa.PrivateKey), true)

	csp, _ := factory.GetSwBccsp(&factory.SwOpts{
		HashFamily: "SHA2",
		SecLevel:   256,
		ByteKeyStore: &factory.ByteKeyStore{
			Value: privateKey,
		},
	})

	return &User{
		ID:          id,
		MspID:       mspID,
		Certificate: []byte(certificate),
		Key:         k,
		Bccsp:       csp,
	}, nil
}

func NewConfigGroup(modPolicy string) *common.ConfigGroup {
	return &common.ConfigGroup{
		Groups:    make(map[string]*common.ConfigGroup),
		Values:    make(map[string]*common.ConfigValue),
		Policies:  make(map[string]*common.ConfigPolicy),
		ModPolicy: modPolicy,
	}
}
