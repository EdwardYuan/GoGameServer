package service_lg

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"gogameserver/lib"
	"gogameserver/service_common"
)

const DefaultPoolSize = 1024

type LoginGate struct {
	*service_common.ServerCommon
	consumeChan chan *sarama.ConsumerMessage
	kafka       *sarama.Client
	workPool    *ants.Pool
}

func NewLoginGate(_name string, id int) *LoginGate {
	eventSvr := gnet.EventServer{}
	client, err := sarama.NewClient(lib.KafkaBroker, sarama.NewConfig())
	pool, err := ants.NewPool(DefaultPoolSize)
	if err != nil {
		fmt.Println(err)
	}
	return &LoginGate{
		ServerCommon: &service_common.ServerCommon{
			Name:        _name,
			Id:          id,
			EventServer: &eventSvr,
			CloseChan:   make(chan int),
		},
		kafka:       &client,
		workPool:    pool,
		consumeChan: make(chan *sarama.ConsumerMessage),
	}
}

func (lg *LoginGate) Close() {

}

func (lg *LoginGate) ConsumeMessage(msg *sarama.ConsumerMessage) {
	//TODO check topic && partition id
	// decode msg and submit task to pool
	lg.workPool.Submit(func() {

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
