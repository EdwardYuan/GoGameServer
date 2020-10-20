package service_gs

import "gogameserver/service_common"

type GameServer struct {
	*service_common.ServerCommon
}

func NewGameServer(name string) *GameServer {
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{},
	}
}

func (gs *GameServer) Start(_name string) (err error) {
	return
}

func (gs *GameServer) Stop() {

}

func (gs *GameServer) Run() {

}
