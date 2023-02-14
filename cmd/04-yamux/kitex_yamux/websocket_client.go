package kitex_yamux

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"nhooyr.io/websocket"
)

type KitexWebsocketYamuxClient[S, C any] struct {
	id                     string
	serverBaseURL          string
	newKitexClientFunc     NewKitexClientFunc[C]
	newKitexServerFunc     NewKitexServerFunc[S]
	kitexServerImplFactory KitexServerImplFactory[S, C]

	Peer    KitexPeer[C]
	s2cConn net.Conn
	c2sConn net.Conn
}

func NewKitexWebsocketYamuxClient[S, C any](id, serverBaseURL string,
	newKitexClientFunc NewKitexClientFunc[C], newKitexServerFunc NewKitexServerFunc[S],
	kitexServerImplFactory KitexServerImplFactory[S, C]) *KitexWebsocketYamuxClient[S, C] {
	return &KitexWebsocketYamuxClient[S, C]{
		id:                     id,
		serverBaseURL:          serverBaseURL,
		newKitexClientFunc:     newKitexClientFunc,
		newKitexServerFunc:     newKitexServerFunc,
		kitexServerImplFactory: kitexServerImplFactory,

		Peer: KitexPeer[C]{
			ID: id,
		},
	}
}

func (c *KitexWebsocketYamuxClient[S, C]) Dial(ctx context.Context) error {
	if err := c.dialS2C(ctx); err != nil {
		return fmt.Errorf("dail s2c error: %s", err)
	}
	if err := c.dialC2S(ctx); err != nil {
		return fmt.Errorf("dail c2s error: %s", err)
	}
	return nil
}

func (c *KitexWebsocketYamuxClient[S, C]) Close(ctx context.Context) error {
	err1, err2 := c.c2sConn.Close(), c.s2cConn.Close()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("c2s error: %s, s2c error: %s", err1, err2)
	}
	return nil
}

func (c *KitexWebsocketYamuxClient[S, C]) dialS2C(ctx context.Context) error {
	// websocket -> net.Conn
	fmt.Println(c.serverBaseURL + "/s2c/" + c.id)
	wsConn, resp, err := websocket.Dial(ctx, c.serverBaseURL+"/s2c/"+c.id, &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
	})
	if err != nil {
		body, err1 := io.ReadAll(resp.Body)
		if err1 != nil {
			body = []byte("")
		}
		return fmt.Errorf("websocket dial error: %s, response body: %s", err, string(body))
	}
	conn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)
	// net.Conn -> yamux server session -> net.Listener -> kitex server
	kitexServer, _, err := NewKitexYamuxServer(conn, &c.Peer, c.newKitexServerFunc, c.kitexServerImplFactory)
	if err != nil {
		_ = conn.Close()
		return err
	}
	go func() {
		err := kitexServer.Run()
		if err != nil {
			// TODO: 预期内的错误处理：
			// EOF
			// failed to get reader: failed to read frame header: EOF
			log.Printf("kitex server run return error: %s", err.Error())
		}
	}()
	c.Peer.Server = kitexServer
	c.s2cConn = conn
	return nil
}

func (c *KitexWebsocketYamuxClient[S, C]) dialC2S(ctx context.Context) error {
	wsConn, resp, err := websocket.Dial(ctx, c.serverBaseURL+"/c2s/"+c.id, &websocket.DialOptions{
		CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
	})
	if err != nil {
		body, err1 := io.ReadAll(resp.Body)
		if err1 != nil {
			body = []byte("")
		}
		return fmt.Errorf("websocket dial error: %s, response body: %s", err, string(body))
	}
	conn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)
	// net.Conn -> yamux kitexClient session -> remote.Dialer -> C
	kitexClient, _, err := NewKitexYamuxClient(conn, c.newKitexClientFunc)
	if err != nil {
		_ = conn.Close()
		return err
	}
	c.Peer.Client = kitexClient
	c.c2sConn = conn
	return nil
}
