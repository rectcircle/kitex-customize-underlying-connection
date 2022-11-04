package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
)

func main() {
	c, err := echo.NewClient("example", client.WithHostPorts("127.0.0.1:8888"))
	if err != nil {
		log.Fatal(err)
	}
	req := &api.Request{Message: "Say hello by netpoll"}
	resp, err := c.Echo(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.Message)
}
