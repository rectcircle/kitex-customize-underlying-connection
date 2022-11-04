package server

import (
	"context"

	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
)

type EchoImpl struct{}

func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	return &api.Response{Message: req.Message}, nil
}
