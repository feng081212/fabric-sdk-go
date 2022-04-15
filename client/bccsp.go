package client

import (
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp"
	"github.com/feng081212/fabric-sdk-go/fabric/bccsp/factory"
)

var defaultBCCSP bccsp.BCCSP

func init()  {
	opts := &factory.SwOpts{
		HashFamily: "SHA2",
		SecLevel:   256,
	}
	csp, err := factory.GetSwBccsp(opts)
	if err != nil {
		panic(err)
	}
	defaultBCCSP = csp
}

func GetDefaultBCCSP() bccsp.BCCSP {
	return defaultBCCSP
}