package kitex_yamux

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"nhooyr.io/websocket"
)

type KitexWebsocketYamuxServer[S, C any] struct {
	server                 *http.Server
	s2cPath                string
	c2sPath                string
	newKitexClientFunc     NewKitexClientFunc[C]
	newKitexServerFunc     NewKitexServerFunc[S]
	kitexServerImplFactory KitexServerImplFactory[S, C]

	peers sync.Map // <clientID, *KitexPeer>
}

func NewKitexWebsocketYamuxServer[S, C any](BindAddr, BasePath string,
	newKitexClientFunc NewKitexClientFunc[C], newKitexServerFunc NewKitexServerFunc[S],
	kitexServerImplFactory KitexServerImplFactory[S, C]) *KitexWebsocketYamuxServer[S, C] {
	mux := http.NewServeMux()
	if !strings.HasPrefix(BasePath, "/") {
		BasePath = "/" + BasePath
	}
	s := &KitexWebsocketYamuxServer[S, C]{
		server:                 &http.Server{Addr: BindAddr, Handler: mux},
		s2cPath:                BasePath + "/s2c/",
		c2sPath:                BasePath + "/c2s/",
		newKitexClientFunc:     newKitexClientFunc,
		newKitexServerFunc:     newKitexServerFunc,
		kitexServerImplFactory: kitexServerImplFactory,
	}
	mux.Handle(s.s2cPath, http.HandlerFunc(s.s2cHandle))
	mux.Handle(s.c2sPath, http.HandlerFunc(s.c2sHandle))
	return s
}

func (s *KitexWebsocketYamuxServer[S, C]) Run() error {
	return s.server.ListenAndServe()
}

func (s *KitexWebsocketYamuxServer[S, C]) s2cHandle(w http.ResponseWriter, r *http.Request) {
	// 获取 clientID
	clientID := strings.TrimPrefix(r.URL.Path, s.s2cPath)
	if clientID == "" {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Bad request, path must is %s<clientID>", s.c2sPath)))
		return
	}
	// 获取 peer 信息
	newP := &KitexPeer[C]{ID: clientID}
	nilClient := newP.Client
	_p, _ := s.peers.LoadOrStore(clientID, newP)
	p := _p.(*KitexPeer[C])
	// 建立 c2s 连接
	closeChan, ok := func() (closeChan <-chan struct{}, ok bool) {
		p.Mutex.Lock()
		defer p.Mutex.Unlock()
		// if p.Client != nil { // 该写法截止 go1.18 不支持，原因是暂未支持 nilable 约束。https://github.com/golang/go/issues/53656
		// if any(p.Client) != nil { // 该写法永远为 true
		if !IsZero(p.Client) { // 只能通过反射判断。
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("Bad request, client %s has establish s2c conn", clientID)))
			return
		}
		// http -> websocekt -> net.Conn
		wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
		})
		if err != nil {
			log.Printf("accept websocket conn error: %v", err)
			return
		}
		conn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)
		// net.Conn -> yamux client session -> remote.Dialer -> C
		kitexClient, closeChan, err := NewKitexYamuxClient(conn, s.newKitexClientFunc)
		if err != nil {
			log.Printf("%v", err)
			return
		}
		// 写入 peer
		p.Client = kitexClient
		ok = true
		return
	}()
	// 阻塞等待连接关闭，并清理。
	if ok {
		<-closeChan
		func() {
			p.Mutex.Lock()
			defer p.Mutex.Unlock()
			p.Client = nilClient
		}()
	}
}

func (s *KitexWebsocketYamuxServer[S, C]) c2sHandle(w http.ResponseWriter, r *http.Request) {
	// 获取 clientID
	clientID := strings.TrimPrefix(r.URL.Path, s.c2sPath)
	if clientID == "" {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Bad request, path must is %s<clientID>", s.c2sPath)))
		return
	}
	// 获取 peer 信息
	_p, _ := s.peers.LoadOrStore(clientID, &KitexPeer[C]{ID: clientID})
	p := _p.(*KitexPeer[C])
	// 建立 c2s 连接
	closeChan, ok := func() (closeChan <-chan struct{}, ok bool) {
		p.Mutex.Lock()
		defer p.Mutex.Unlock()
		if p.Server != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("Bad request, client %s has establish c2s conn", clientID)))
			return
		}
		// if p.Client == nil { // 该写法截止 go1.18 不支持，原因是暂未支持 nilable 约束。https://github.com/golang/go/issues/53656
		// if any(p.Client) == nil { // 该写法永远为 false
		if IsZero(p.Client) { // 只能通过反射判断。
			w.WriteHeader(400)
			_, _ = w.Write([]byte("Bad request, must first establish s2c conn"))
			return
		}
		// http -> websocekt -> net.Conn
		wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			CompressionMode: websocket.CompressionDisabled, // 默认压缩模式有概率触发 panic，禁用之。
		})
		if err != nil {
			log.Printf("accept websocket conn error: %v", err)
			return
		}
		conn := websocket.NetConn(context.Background(), wsConn, websocket.MessageBinary)
		// net.Conn -> yamux server session -> net.Listener -> kitex server
		kitexServer, closeChan, err := NewKitexYamuxServer(conn, p, s.newKitexServerFunc, s.kitexServerImplFactory)
		if err != nil {
			log.Printf("%v", err)
			return
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
		// 写入 peer
		p.Server = kitexServer
		// 构造返回值
		ok = true
		return
	}()
	// 阻塞等待连接关闭。
	if ok {
		<-closeChan
		func() {
			p.Mutex.Lock()
			defer p.Mutex.Unlock()
			// if p.Server != nil {
			_ = p.Server.Stop()
			// }
			p.Server = nil
		}()
	}
}
