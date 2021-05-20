package service_gs

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/lib"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
)

type GameServer struct {
	*service_common.ServerCommon
	workPool     *ants.Pool
	protoFactory *protocol.Factory
}

func NewGameServer(_name string) *GameServer {
	pool, err1 := ants.NewPool(1024)
	if err1 != nil {
		fmt.Println("NewGameServer Error: ", err1)
		return nil
	}
	lib.InitLogger()
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
		},
		workPool: pool,
	}
}

func (gs *GameServer) Start() (err error) {
	gnet.Serve(gs, lib.GNetAddr, gnet.WithMulticore(true), gnet.WithCodec(&lib.MsgCodec{}))
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	return
}

func (gs *GameServer) Stop() {
	lib.SugarLogger.Info("Service ", gs.Name, " Stopped.")
}

func (gs *GameServer) Run() {
	lib.SugarLogger.Info("Service run")

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
