package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
)

func main() {
	c, err := echo.NewClient("example",
		client.WithHostPorts("127.0.0.1:8889"),
		// 改造点：client 传输层使用 go 标准网络库
		client.WithTransHandlerFactory(gonet.NewCliTransHandlerFactory()),
	)
	if err != nil {
		log.Fatal(err)
	}
	req := &api.Request{Message: "Say hello by go std net"}
	resp, err := c.Echo(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.Message)
}
