package endpoints

import (
	logging "github.com/feng081212/fabric-sdk-go/common/logger"
)

var logger = logging.NewLogger("fabsdk/endpoints")

const (
	// GRPC max message size (same as Fabric)
	maxCallRecvMsgSize = 100 * 1024 * 1024
	maxCallSendMsgSize = 100 * 1024 * 1024
)