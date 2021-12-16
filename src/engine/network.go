package engine

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/lib"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"sync"
	"time"
)

type Network struct {
	gnet.EventHandler
	socks sync.Map
}

func NewNetwork() *Network {
	return &Network{
		socks: sync.Map{},
	}
}

type EventPool struct {
	pool ants.Pool
}

func delSock(_, conn interface{}) bool {
	err := conn.(gnet.Conn).Close()
	return err == nil
}

func (n *Network) CloseAll() {
	n.socks.Range(delSock)
}

func (n *Network) Start(addr string) error {
	err := gnet.Serve(n, addr, gnet.WithMulticore(true),
		gnet.WithCodec(codec.CodecProtobuf{}),
		gnet.WithLogger(lib.SugarLogger))
	return err
}

func (n *Network) OnInitComplete(server gnet.Server) (action gnet.Action) {
	return
}

func (n *Network) OnShutdown(server gnet.Server) {
	return
}

func (n *Network) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	n.socks.Store(c.RemoteAddr(), c)
	return
}

func (n *Network) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	n.socks.Delete(c)
	return
}

func (n *Network) PreWrite() {

}

func (n *Network) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return
}

func (n *Network) Tick() (delay time.Duration, action gnet.Action) {
	return
}
