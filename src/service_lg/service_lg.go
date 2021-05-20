package service_lg

import (
	"GoGameServer/src/service_common"
	"github.com/Shopify/sarama"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
)

const DefaultPoolSize = 1024

var ServiceLogin *LoginGate

func init() {
	ServiceLogin = NewLoginGate("loginGate", 0)
}

type LoginGate struct {
	*service_common.ServerCommon
	workPool *ants.Pool
	err      error
}

func NewLoginGate(_name string, id int) *LoginGate {
	eventSvr := gnet.EventServer{}
	pool, err := ants.NewPool(DefaultPoolSize)
	service_common.FailOnError(err, "LoginGate Make Pool Error")
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

func (lg *LoginGate) ConsumeMessage(msg *sarama.ConsumerMessage) {
	//TODO check topic && partition id
	// decode msg and submit task to pool
	lg.err = lg.workPool.Submit(func() {

	})
}

func (lg *LoginGate) run() {
	for {
		select {
		case msg, ok := <-lg.consumeChan:
			if ok {
				lg.ConsumeMessage(msg)
			}
		case <-lg.CloseChan:
			lg.Close()
		}
	}
}
