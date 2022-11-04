package main

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/cloudwego/kitex/server"
	api "github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
	serverImpl "github.com/rectcircle/kitex-customize-underlying-connection/server"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", ":8889")
	if err != nil {
		panic(err)
	}
	svr := api.NewServer(new(serverImpl.EchoImpl),
		server.WithServiceAddr(addr),
		// 改造点：server 传输层使用 go 标准网络库
		server.WithTransServerFactory(gonet.NewTransServerFactory()),
		server.WithTransHandlerFactory(gonet.NewSvrTransHandlerFactory()),
	)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
