package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rectcircle/kitex-customize-underlying-connection/cmd/04-yamux/kitex_yamux"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api"
	"github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
)

type ClientEchoImpl struct {
	peer *kitex_yamux.KitexPeer[echo.Client]
}

func NewClientEchoImpl(peer *kitex_yamux.KitexPeer[echo.Client]) api.Echo {
	return &ClientEchoImpl{
		peer: peer,
	}
}

// Echo implements api.Echo
func (c *ClientEchoImpl) Echo(ctx context.Context, req *api.Request) (r *api.Response, err error) {
	log.Printf("client receive message: %s (clientID=%s)", req.Message, c.peer.ID)
	return &api.Response{Message: req.Message}, nil
}

// go run ./cmd/04-yamux/client 0
// go run ./cmd/04-yamux/client 1
func main() {
	fmt.Println(strings.Join(os.Args, " "))
	c := kitex_yamux.NewKitexWebsocketYamuxClient(os.Args[1], "http://127.0.0.1:8891/kitex-ws", echo.NewClient, echo.NewServer, NewClientEchoImpl)
	err := c.Dial(context.Background())
	if err != nil {
		log.Println(err.Error())
		return
	}

	callResp, err := c.Peer.Client.Echo(context.Background(), &api.Request{Message: "client call server"})
	if err != nil {
		log.Printf("client call server error: %s", err)
		return
	}
	log.Printf("client call server response: %s (clientID=%s)", callResp.Message, c.Peer.ID)
	time.Sleep(1 * time.Second)
	c.Close(context.Background())
}
