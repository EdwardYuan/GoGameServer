package service_gs

import (
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"gogameserver/MsgHandler"
	"gogameserver/service_common"
)

type GameServer struct {
	*service_common.ServerCommon
	workPool *ants.Pool
	name     string
}

func NewGameServer(_name string) *GameServer {
	gs := &GameServer{
		ServerCommon: &service_common.ServerCommon{},
		name:         _name,
	}
	pool, err := ants.NewPool(1024)
	if err == nil {
		gs.workPool = pool
	}
	return gs
}

func (gs *GameServer) Start() (err error) {
	gnet.Serve(gs, "tcp://127.0.0.1:9000", gnet.WithMulticore(true))
	return
}

func (gs *GameServer) Stop() {

}

func (gs *GameServer) Run() {

}

func (gs *GameServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg, err := gs.Decode(frame)
	if err != nil {
		gs.AddMessageNode(&msg)
	}
	return
}

func (gs *GameServer) AddMessageNode(msg *MsgHandler.Message) {
	gs.workPool.Submit(func() {

	})

}
