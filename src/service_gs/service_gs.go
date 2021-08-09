package service_gs

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"sync"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	wg           sync.WaitGroup
	workPool     *ants.Pool
	protoFactory *protocol.Factory
}

func NewGameServer(_name string, id int) *GameServer {
	pool, err1 := ants.NewPool(global.DefaultPoolSize)
	lib.FatalOnError(err1, "NewGameServer Error")
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		wg:         sync.WaitGroup{},
		runChannel: make(chan bool),
		workPool:   pool,
	}
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	err = gnet.Serve(gs, lib.GNetAddr, gnet.WithMulticore(true), gnet.WithCodec(&lib.MsgCodec{}),
		gnet.WithLogger(lib.SugarLogger))
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	gs.Run()
	return
}

func (gs *GameServer) Stop() {
	defer func() {
		gs.workPool.Release()
		close(gs.runChannel)
	}()
	gs.wg.Wait()
	gs.runChannel <- false
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
	gs.wg.Add(1)
	gs.workPool.Submit(func() {
		defer gs.wg.Done()

	})
}
