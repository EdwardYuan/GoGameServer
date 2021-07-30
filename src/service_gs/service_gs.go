package service_gs

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/lib"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	workPool     *ants.Pool
	protoFactory *protocol.Factory
}

func NewGameServer(_name string, id int) *GameServer {
	pool, err1 := ants.NewPool(1024)
	lib.FatalOnError(err1, "NewGameServer Error")
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		runChannel: make(chan bool),
		workPool:   pool,
	}
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	go gnet.Serve(gs, lib.GNetAddr, gnet.WithMulticore(true), gnet.WithCodec(&lib.MsgCodec{}))
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	gs.Run()
	return
}

func (gs *GameServer) Stop() {
	gs.runChannel <- false
	close(gs.runChannel)
	lib.SugarLogger.Info("Service ", gs.Name, " Stopped.")
}

func (gs *GameServer) Run() {
	for {
		select {
		case <-gs.CloseChan:
			gs.Stop()
		case <-gs.runChannel:
			lib.SugarLogger.Info("running...")
		default:
		}
	}
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
