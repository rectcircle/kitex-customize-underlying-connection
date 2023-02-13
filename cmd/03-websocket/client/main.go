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
	"nhooyr.io/websocket"
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
	wsConn, _, err := websocket.Dial(context.Background(), d.ServerURL, &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
	})
	if err != nil {
		return nil, err
	}
	return websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary), nil
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
