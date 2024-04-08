package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GrpcStatus struct {
	Code int
}

func (c GrpcStatus) AddStatus(ctx *context.Context) {
	_ = grpc.SetHeader(*ctx, metadata.Pairs("x-http-code", "401"))
}
