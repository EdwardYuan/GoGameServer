package service_gs

import (
	"github.com/Shopify/sarama"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"gogameserver/MsgHandler"
	"gogameserver/lib"
	"gogameserver/service_common"
)

type GameServer struct {
	*service_common.ServerCommon
	workPool    *ants.Pool
	name        string
	KafkaClient sarama.Client
}

func NewGameServer(_name string) *GameServer {
	gs := &GameServer{
		ServerCommon: &service_common.ServerCommon{},
		name:         _name,
	}
	lib.InitLogger()
	lib.SugarLogger.Info("Service ", _name, " created")
	pool, err := ants.NewPool(1024)
	if err == nil {
		gs.workPool = pool
	}
	gs.KafkaClient, err = sarama.NewClient(lib.KafkaBroker, sarama.NewConfig())
	if err != nil {
		lib.SugarLogger.Error(err)
	}
	return gs
}

func (gs *GameServer) Start() (err error) {
	gnet.Serve(gs, "tcp://127.0.0.1:9000", gnet.WithMulticore(true))
	lib.SugarLogger.Info("Service ", gs.name, " Start...")
	return
}

func (gs *GameServer) Stop() {
	lib.SugarLogger.Info("Service ", gs.name, " Stopped.")
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

func (gs *GameServer) NewConsumer(topic string) (consumer sarama.Consumer, err error) {
	return sarama.NewConsumer(lib.KafkaBroker, gs.KafkaClient.Config())
}
