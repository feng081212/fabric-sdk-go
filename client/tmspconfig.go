package client

import (
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/hasher"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/sw"
	"github.com/feng081212/fabric-sdk-go/fabric/crypto/certs"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type OUIdentifier struct {
	OrganizationUnitIdentifier string
	Certificate                string
}

type NodeOUs struct {
	Enable                  bool
	ClientOUIdentifier      string
	ClientOUIdentifierCert  string
	PeerOUIdentifier        string
	PeerOUIdentifierCert    string
	AdminOUIdentifier       string
	AdminOUIdentifierCert   string
	OrdererOUIdentifier     string
	OrdererOUIdentifierCert string
}

type MspConfig struct {
	MspID                string
	Cacerts              []string
	Intermediatecerts    []string
	AdminCerts           []string
	TlsCACerts           []string
	TlsIntermediateCerts []string
	Crls                 []string
	NodeOUs              *NodeOUs
	OUIdentifiers        []*OUIdentifier
}

func (p *MspConfig) BuildMspConfig() (*msp.MSPConfig, error) {
	cacerts, e := certs.ReadPemMaterialFromStrings(p.Cacerts)
	if e != nil {
		return nil, fmt.Errorf("load cacerts error with msp[%s] : %v", p.MspID, e)
	}

	admincert, e := certs.ReadPemMaterialFromStrings(p.AdminCerts)
	if e != nil {
		return nil, fmt.Errorf("load admincert error with msp[%s] : %v", p.MspID, e)
	}

	intermediatecerts, e := certs.ReadPemMaterialFromStrings(p.Intermediatecerts)
	if e != nil {
		return nil, fmt.Errorf("load intermediatecerts error with msp[%s] : %v", p.MspID, e)
	}

	crls, e := certs.ReadPemMaterialFromStrings(p.Crls)
	if e != nil {
		return nil, fmt.Errorf("load crls error with msp[%s] : %v", p.MspID, e)
	}

	tlsCACerts, e := certs.ReadPemMaterialFromStrings(p.TlsCACerts)
	if e != nil {
		return nil, fmt.Errorf("load tlsCACerts error with msp[%s] : %v", p.MspID, e)
	}

	tlsIntermediateCerts, e := certs.ReadPemMaterialFromStrings(p.TlsIntermediateCerts)
	if e != nil {
		return nil, fmt.Errorf("load tlsIntermediateCerts error with msp[%s] : %v", p.MspID, e)
	}

	var ouIdentifiers []*msp.FabricOUIdentifier
	nodeOUs := &msp.FabricNodeOUs{}

	if p.OUIdentifiers != nil && len(p.OUIdentifiers) > 0 {
		for _, ouIdentifier := range p.OUIdentifiers {
			cert, e := certs.ReadPemFromString(ouIdentifier.Certificate)
			if e != nil {
				return nil, fmt.Errorf("load ouid [%s] error with msp[%s] : %v", ouIdentifier.OrganizationUnitIdentifier, p.MspID, e)
			}
			oui := &msp.FabricOUIdentifier{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: ouIdentifier.OrganizationUnitIdentifier,
			}
			ouIdentifiers = append(ouIdentifiers, oui)
		}
	}

	if p.NodeOUs != nil && p.NodeOUs.Enable {
		nodeOUs.Enable = true

		if p.NodeOUs.ClientOUIdentifier == "" {
			p.NodeOUs.ClientOUIdentifier = "client"
		}
		clientCaCert, e := certs.ReadPemFromString(p.NodeOUs.ClientOUIdentifierCert)
		if e == nil {
			nodeOUs.ClientOuIdentifier = &msp.FabricOUIdentifier{
				Certificate:                  clientCaCert,
				OrganizationalUnitIdentifier: p.NodeOUs.ClientOUIdentifier,
			}
		}

		if p.NodeOUs.PeerOUIdentifier == "" {
			p.NodeOUs.PeerOUIdentifier = "peer"
		}
		peerCaCert, e := certs.ReadPemFromString(p.NodeOUs.PeerOUIdentifierCert)
		if e == nil {
			nodeOUs.PeerOuIdentifier = &msp.FabricOUIdentifier{
				Certificate:                  peerCaCert,
				OrganizationalUnitIdentifier: p.NodeOUs.PeerOUIdentifier,
			}
		}

		if p.NodeOUs.AdminOUIdentifier == "" {
			p.NodeOUs.AdminOUIdentifier = "admin"
		}
		adminCaCert, e := certs.ReadPemFromString(p.NodeOUs.AdminOUIdentifierCert)
		if e == nil {
			nodeOUs.AdminOuIdentifier = &msp.FabricOUIdentifier{
				Certificate:                  adminCaCert,
				OrganizationalUnitIdentifier: p.NodeOUs.AdminOUIdentifier,
			}
		}

		if p.NodeOUs.OrdererOUIdentifier == "" {
			p.NodeOUs.OrdererOUIdentifier = "orderer"
		}
		ordererCaCert, e := certs.ReadPemFromString(p.NodeOUs.OrdererOUIdentifierCert)
		if e == nil {
			nodeOUs.OrdererOuIdentifier = &msp.FabricOUIdentifier{
				Certificate:                  ordererCaCert,
				OrganizationalUnitIdentifier: p.NodeOUs.OrdererOUIdentifier,
			}
		}
	}

	// Set FabricCryptoConfig
	cryptoConfig := &msp.FabricCryptoConfig{
		SignatureHashFamily:            sw.SHA2,
		IdentityIdentifierHashFunction: string(hasher.SHA256),
	}

	// Compose FabricMSPConfig
	fabricMspConf := &msp.FabricMSPConfig{
		Admins:                        admincert,
		RootCerts:                     cacerts,
		IntermediateCerts:             intermediatecerts,
		SigningIdentity:               nil,
		Name:                          p.MspID,
		OrganizationalUnitIdentifiers: ouIdentifiers,
		RevocationList:                crls,
		CryptoConfig:                  cryptoConfig,
		TlsRootCerts:                  tlsCACerts,
		TlsIntermediateCerts:          tlsIntermediateCerts,
		FabricNodeOus:                 nodeOUs,
	}

	fabricMpsJs, err := proto.Marshal(fabricMspConf)
	if err != nil {
		return nil, err
	}

	return &msp.MSPConfig{Config: fabricMpsJs, Type: 0}, nil
}
