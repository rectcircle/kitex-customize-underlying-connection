package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
	"golang.org/x/net/websocket"
)

type WebsocketKitexDialer struct {
	ServerURL string
}

func NewWebsocketKitexDialer(serverURL string) remote.Dialer {
	return &WebsocketKitexDialer{
		ServerURL: serverURL,
	}
}

// DialTimeout implements remote.Dialer
func (d *WebsocketKitexDialer) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	cfg, err := websocket.NewConfig(d.ServerURL, d.ServerURL)
	if err != nil {
		return nil, err
	}
	return websocket.DialConfig(cfg)
}

func main() {
	c, err := echo.NewClient("example",
		// 这只是一个 mock
		client.WithHostPorts("127.0.0.1:8890"),
		// 改造点：使用自定义 dialer 获取 net.Conn
		client.WithDialer(NewWebsocketKitexDialer("ws://127.0.0.1:8890/kitex-ws")),
		// 改造点：client 传输层使用 go 标准网络库
		client.WithTransHandlerFactory(gonet.NewCliTransHandlerFactory()),
	)
	if err != nil {
		log.Fatal(err)
	}
	req := &api.Request{Message: "Say hello by websocket"}
	resp, err := c.Echo(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp.Message)
}
