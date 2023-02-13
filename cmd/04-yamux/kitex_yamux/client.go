package kitex_yamux

import (
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/remote"
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
