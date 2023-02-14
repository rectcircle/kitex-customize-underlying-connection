package kitex_yamux

import (
	"fmt"
	"net"

	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/cloudwego/kitex/server"
	"github.com/hashicorp/yamux"
	"github.com/rectcircle/kitex-customize-underlying-connection/cmd/04-yamux/kitex_yamux/kitexfixed"
)

type kitexYamuxListener struct {
	serverSession *yamux.Session
}

func NewKitexYamuxListener(serverSession *yamux.Session) net.Listener {
	return &kitexYamuxListener{
		serverSession: serverSession,
	}
}

// Accept implements net.Listener
func (l *kitexYamuxListener) Accept() (net.Conn, error) {
	return l.serverSession.Accept()
}

// Addr implements net.Listener
func (l *kitexYamuxListener) Addr() net.Addr {
	return l.serverSession.Addr()
}

// Close implements net.Listener
func (l *kitexYamuxListener) Close() error {
	return l.serverSession.Close()
}

type NewKitexServerFunc[S any] func(handler S, opts ...server.Option) server.Server
type KitexServerImplFactory[S, C any] func(peer *KitexPeer[C]) S

func NewKitexYamuxServer[S, C any](conn net.Conn, p *KitexPeer[C], newKitexServerFunc NewKitexServerFunc[S], kitexServerImplFactory KitexServerImplFactory[S, C]) (s server.Server, closeChan <-chan struct{}, err error) {
	// net.Conn -> yamux server session -> net.Listener
	serverSession, err := yamux.Server(conn, nil)
	if err != nil {
		err = fmt.Errorf("creat yamux server session error: %v", err)
		return
	}
	serverListener := NewKitexYamuxListener(serverSession)
	// net.Listener -> kitex server
	s = newKitexServerFunc(kitexServerImplFactory(p),
		server.WithListener(serverListener),
		// 改造点：server 传输层使用 go 标准网络库
		// github.com/cloudwego/kitex@v0.4.5-0.20230213035731-7054d09a7d3a/pkg/remote/trans/gonet/trans_server.go:81
		// kitex accept 错误处理存在问题，直接 os.Exit，这是个坏的设计。这里不得以，自己修改了源码，放到了 kitexfixed 回避一下其问题。
		// server.WithTransServerFactory(gonet.NewTransServerFactory()),
		server.WithTransServerFactory(kitexfixed.NewTransServerFactory()),
		server.WithTransHandlerFactory(gonet.NewSvrTransHandlerFactory()),
	)
	closeChan = serverSession.CloseChan()
	return
}
