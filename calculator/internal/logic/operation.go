package logic

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func overflowError() error {
	return status.Error(codes.OutOfRange, "integer overflow")
}
