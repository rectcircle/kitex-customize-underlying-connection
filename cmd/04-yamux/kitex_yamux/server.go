package kitex_yamux

import (
	"net"

	"github.com/hashicorp/yamux"
)

type kitexYamuxListener struct {
	serverSession *yamux.Session
}

func NewKitexYamuxListener(clientSession *yamux.Session) net.Listener {
	return &kitexYamuxListener{
		serverSession: clientSession,
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
