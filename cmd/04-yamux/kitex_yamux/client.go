package kitex_yamux

import (
	"fmt"
	"net"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/hashicorp/yamux"
)

type kitexYamuxDialer struct {
	clientSession *yamux.Session
}

func NewKitexYamuxDialer(clientSession *yamux.Session) remote.Dialer {
	return &kitexYamuxDialer{
		clientSession: clientSession,
	}
}

// DialTimeout implements remote.Dialer
func (d *kitexYamuxDialer) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	return d.clientSession.Open()
}

type NewKitexClientFunc[C any] func(destService string, opts ...client.Option) (C, error)

func NewKitexYamuxClient[C any](conn net.Conn, newKitexClientFunc NewKitexClientFunc[C]) (c C, closeChan <-chan struct{}, err error) {
	// net.Conn -> yamux client session -> remote.Dialer
	clientSession, err := yamux.Client(conn, nil)
	if err != nil {
		err = fmt.Errorf("creat yamux client session error: %v", err)
		return
	}
	cleintDialer := NewKitexYamuxDialer(clientSession)
	// remote.Dialer -> C
	c, err = newKitexClientFunc("yamux",
		// 这只是一个 mock
		client.WithHostPorts("127.0.0.1:65535"),
		// 改造点：使用自定义 dialer 获取 net.Conn
		client.WithDialer(cleintDialer),
		// 改造点：client 传输层使用 go 标准网络库
		client.WithTransHandlerFactory(gonet.NewCliTransHandlerFactory()),
	)
	if err != nil {
		err = fmt.Errorf("creat kitex client error: %v", err)
		return
	}
	closeChan = clientSession.CloseChan()
	return
}
