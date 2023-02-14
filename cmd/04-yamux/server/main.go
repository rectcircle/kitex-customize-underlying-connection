package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rectcircle/kitex-customize-underlying-connection/cmd/04-yamux/kitex_yamux"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
)

type ServerEchoImpl struct {
	peer *kitex_yamux.KitexPeer[echo.Client]
}

func NewServerEchoImpl(peer *kitex_yamux.KitexPeer[echo.Client]) api.Echo {
	return &ServerEchoImpl{
		peer: peer,
	}
}

func (s *ServerEchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	log.Printf("server receive message: %s (clientID=%s)", req.Message, s.peer.ID)
	// server 主动发消息给 client
	callResp, err := s.peer.Client.Echo(ctx, &api.Request{Message: "server call client"})
	if err != nil {
		log.Printf("server call client error: %s", err)
		return nil, fmt.Errorf("server call client error: %s", err)
	}
	log.Printf("server call client response: %s (clientID=%s)", callResp.Message, s.peer.ID)
	// 返回
	return &api.Response{Message: req.Message}, nil
}

// go run ./cmd/04-yamux/server
func main() {

	// ws://:8891/kitex-ws/c2s/<client-id>
	// ws://:8891/kitex-ws/s2c/<client-id>
	s := kitex_yamux.NewKitexWebsocketYamuxServer(":8891", "/kitex-ws", echo.NewClient, echo.NewServer, NewServerEchoImpl)
	log.Printf("websocket listen on :8891/kitex-ws")
	err := s.Run()
	if err != nil {
		log.Println(err.Error())
	}
	// TODO 两个 ctrl + c 才能停止问题。
}
