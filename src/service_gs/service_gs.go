package service_gs

import (
	"github.com/panjf2000/gnet"
	"gogameserver/service_common"
)

type GameServer struct {
	*service_common.ServerCommon
	name string
}

func NewGameServer(_name string) *GameServer {
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{},
		name:         _name,
	}
}

func (gs *GameServer) Start() (err error) {
	gnet.Serve(gs, "tcp://127.0.0.1:9000", gnet.WithMulticore(true))
	return
}

func (gs *GameServer) Stop() {

}

func (gs *GameServer) Run() {

}
