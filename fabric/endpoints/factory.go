package endpoints

import (
	"google.golang.org/grpc/keepalive"
	"time"
)



func EmptyPeer() *Peer {
	return &Peer{
		keepaliveParams: GetDefaultKeepaliveParams(),
		failFast:        false,
		inSecure:        false,
	}
}

func EmptyOrderer() *Orderer {
	return &Orderer{
		keepaliveParams: GetDefaultKeepaliveParams(),
		failFast:        false,
		allowInsecure:   false,
	}
}

func GetDefaultKeepaliveParams() *keepalive.ClientParameters {
	return &keepalive.ClientParameters{
		Time:                0,
		Timeout:             time.Second * 120,
		PermitWithoutStream: false,
	}
}
