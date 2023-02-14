package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/cloudwego/kitex/server"
	api "github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
	serverImpl "github.com/rectcircle/kitex-customize-underlying-connection/server"
	"nhooyr.io/websocket"
)

type WebsocketAddr struct {
	URL *url.URL
}

func ResolveWebsocketAddr(rawURL string) (*WebsocketAddr, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &WebsocketAddr{
		URL: u,
	}, nil
}

// Network implements net.Addr
func (a *WebsocketAddr) Network() string {
	return a.URL.Scheme
}

// String implements net.Addr
func (a *WebsocketAddr) String() string {
	return strings.TrimPrefix(a.URL.String(), a.URL.Scheme+"://")
}

type ClosedConnWrapper struct {
	net.Conn
	closed        chan struct{}
	closeChanOnce sync.Once
}

func NewClosedConnWrapper(c net.Conn) *ClosedConnWrapper {
	return &ClosedConnWrapper{
		Conn:   c,
		closed: make(chan struct{}),
	}
}

func (c *ClosedConnWrapper) Close() error {
	// fmt.Println("=====")
	c.closeChanOnce.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

func (c *ClosedConnWrapper) CloseChan() <-chan struct{} {
	return c.closed
}

type WebsocketKitexServer struct {
	addr     *WebsocketAddr
	server   *http.Server
	connChan chan net.Conn
}

func NewWebsocketKitexServer(websocketURL string) (*WebsocketKitexServer, error) {
	a, err := ResolveWebsocketAddr(websocketURL)
	if err != nil {
		return nil, err
	}
	return &WebsocketKitexServer{
		addr:     a,
		connChan: make(chan net.Conn),
	}, nil
}

// Accept implements net.Listener
func (s *WebsocketKitexServer) Accept() (net.Conn, error) {
	return <-s.connChan, nil
}

// Addr implements net.Listener
func (s *WebsocketKitexServer) Addr() net.Addr {
	return s.addr
}

// Close implements net.Listener
func (s *WebsocketKitexServer) Close() error {
	return s.server.Close()
}

func (s *WebsocketKitexServer) websocketHandle(w http.ResponseWriter, r *http.Request) {
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
	})
	if err != nil {
		log.Printf("accept websocket conn error: %v", err)
		return
	}
	c := NewClosedConnWrapper(websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary))
	s.connChan <- c
	<-c.CloseChan()
}

func (s *WebsocketKitexServer) Start() error {
	mux := http.NewServeMux()
	mux.Handle(s.addr.URL.Path, http.HandlerFunc(s.websocketHandle))

	server := &http.Server{Addr: s.addr.URL.Host, Handler: mux}
	go server.ListenAndServe() // nolint
	s.server = server
	return nil
}

func main() {
	l, err := NewWebsocketKitexServer("ws://[::]:8890/kitex-ws")
	if err != nil {
		panic(err)
	}
	l.Start()
	svr := api.NewServer(new(serverImpl.EchoImpl),
		server.WithListener(l),
		// 改造点：server 传输层使用 go 标准网络库
		server.WithTransServerFactory(gonet.NewTransServerFactory()),
		server.WithTransHandlerFactory(gonet.NewSvrTransHandlerFactory()),
	)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
