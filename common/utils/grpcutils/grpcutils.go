package grpcutils

import (
	"context"
	logging "github.com/feng081212/fabric-sdk-go/common/logger"
	"google.golang.org/grpc"
)

var logger = logging.NewLogger("fabsdk/grpc/utils")

func DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	logger.Debugf("DialContext [%s]", target)
	opts = append(opts, grpc.WithBlock())
	return grpc.DialContext(ctx, target, opts...)
}

func ReleaseConn(conn *grpc.ClientConn) {
	logger.Debugf("ReleaseConn [%p]", conn)
	if err := conn.Close(); err != nil {
		logger.Debugf("unable to close connection [%s]", err)
	}
}
