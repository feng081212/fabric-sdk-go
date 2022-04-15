package endpoints

import (
	status2 "fabric-sdk-go/fabric/errors/status"
	"fmt"
)

func ParseGrpcError(e error, g status2.Group, message ...string) error {
	s := status2.NewFromGRPCError(e)
	if s != nil {
		return s
	}
	msg := e.Error()
	if len(message) > 0 {
		msg = fmt.Sprintf("%s : %s", message[0], msg)
	}
	return status2.New(g, status2.ConnectionFailed.ToInt32(), msg)
}
