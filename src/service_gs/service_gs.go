package service_gs

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
	"sync"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	wg sync.WaitGroup
}

func NewGameServer(_name string, id int) *GameServer {
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		wg:         sync.WaitGroup{},
		runChannel: make(chan bool),
	}
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	gs.Run()
	return
}

func (gs *GameServer) Stop() {
	defer func() {
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
