package endpoints

import (
	"fmt"
	status2 "github.com/feng081212/fabric-sdk-go/fabric/errors/status"
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
