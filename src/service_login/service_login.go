package service_login

import (
	"encoding/json"
	"fmt"
	"sync"

	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"

	"github.com/panjf2000/ants/v2"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

type LoginGate struct {
	*service_common.ServerCommon
	workPool   *ants.Pool
	Rabbit     *lib.RabbitClient
	err        error
	deliveries <-chan amqp.Delivery
	stopOnce   sync.Once
}

func NewLoginGate(_name string, id int) *LoginGate {
	pool, err := ants.NewPool(global.DefaultPoolSize)
	lib.FatalOnError(err, "LoginGate Make Pool Error")
	return &LoginGate{
		ServerCommon: service_common.NewServerCommon(_name, id),
		workPool:     pool,
		Rabbit:       lib.NewRabbitClient(),
	}
}

func (lg *LoginGate) Error() string {
	if lg.err != nil {
		lib.SugarLogger.Error(lg.err)
		return "LoginGate Error"
	}
	return ""
}

func (lg *LoginGate) Stop() {
	lg.stopOnce.Do(func() {
		lg.ServerCommon.Stop()
		if lg.Rabbit != nil {
			lg.Rabbit.Stop()
		}
		if lg.workPool != nil {
			lg.workPool.Release()
		}
		lib.SugarLogger.Infof("LoginGate %d Closed.", lg.Id)
	})
}

func (lg *LoginGate) Start() (err error) {
	if err = lg.ServerCommon.Start(); err != nil {
		return err
	}
	if lg.Rabbit == nil {
		lg.Rabbit = lib.NewRabbitClient()
	}
	err = lg.Rabbit.Start(config.RabbitUrl, "exchange", "queue", "fanout")
	if err != nil {
		return err
	}
	lg.deliveries, err = lg.Rabbit.SimpleConsume("queue", "")
	if err != nil {
		return err
	}
	// conn, err := net.Dial("tcp", config.GameServerAddr+config.GameServerPort)
	// lib.FatalOnError(err, "logingate connect to gameserver error")
	// if conn != nil {
	//	conn.Write([]byte("Hello GameServer."))
	// }
	// defer conn.Close()
	go lg.Run()
	return err
}

func (lg *LoginGate) Run() {
	for {
		select {
		case msg, ok := <-lg.deliveries:
			if !ok {
				lg.Stop()
				return
			}
			lg.handleDelivery(msg)
		case <-lg.CloseChan:
			lg.Stop()
			return
		}
	}
}

func (lg *LoginGate) LoadConfig(path string) error {
	return lg.ServerCommon.LoadConfig(path)
}

func (lg *LoginGate) handleDelivery(msg amqp.Delivery) {
	switch msg.ContentType {
	case "json":
		type Person struct {
			Id    int
			Name  string
			Email string
		}
		var p Person
		if err := json.Unmarshal(msg.Body, &p); err != nil {
			lib.Log(zapcore.DebugLevel, "json unmarshal msg error", err)
		}
		lib.Log(zapcore.DebugLevel, fmt.Sprintln(p), nil)
	case "protobuf":
		var p1 pb.Person
		if err := proto.Unmarshal(msg.Body, &p1); err != nil {
			lib.Log(zapcore.DebugLevel, "proto unmarshal msg error", err)
		}
		lib.Log(zapcore.DebugLevel, fmt.Sprintln(&p1), nil)
	default:
		lib.Log(zap.DebugLevel, "nothing", nil)
	}
	if err := msg.Ack(false); err != nil {
		lib.LogIfError(err, "LoginGate ack message error")
	}
}
