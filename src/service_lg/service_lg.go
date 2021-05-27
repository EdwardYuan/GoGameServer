package service_lg

import (
	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
)

const DefaultPoolSize = 1024

var ServiceLogin *LoginGate

type LoginGate struct {
	*service_common.ServerCommon
	workPool *ants.Pool
	err      error
}

func NewLoginGate(_name string, id int) *LoginGate {
	eventSvr := gnet.EventServer{}
	pool, err := ants.NewPool(DefaultPoolSize)
	lib.FatalOnError(err, "LoginGate Make Pool Error")
	global.ServerMap.MapAddrToServerName(config.LoginGateAddr, global.ServerLogin)
	return &LoginGate{
		ServerCommon: &service_common.ServerCommon{
			Name:        _name,
			Id:          id,
			EventServer: &eventSvr,
			CloseChan:   make(chan int),
		},
		workPool: pool,
	}
}

func (lg *LoginGate) Close() {

}

func (lg *LoginGate) Start() {
	lg.run()
}

func (lg *LoginGate) run() {
	select {
	case <-lg.CloseChan:
		lg.Close()
	}
}
