package kitex_yamux

import (
	"sync"

	"github.com/cloudwego/kitex/server"
)

type KitexPeer[C any] struct {
	Mutex  sync.Mutex
	ID     string
	Client C
	Server server.Server
}
