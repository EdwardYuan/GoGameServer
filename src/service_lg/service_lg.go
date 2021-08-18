package service_lg

import (
	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"encoding/json"
	"fmt"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

type LoginGate struct {
	*service_common.ServerCommon
	workPool *ants.Pool
	Rabbit   *lib.RabbitClient
	err      error
}

func NewLoginGate(_name string, id int) *LoginGate {
	pool, err := ants.NewPool(global.DefaultPoolSize)
	lib.FatalOnError(err, "LoginGate Make Pool Error")
	return &LoginGate{
		ServerCommon: &service_common.ServerCommon{
			Name:      _name,
			Id:        id,
			CloseChan: make(chan int),
		},
		workPool: pool,
		Rabbit:   lib.NewRabbitClient(),
	}
}

func (lg *LoginGate) Stop() {
	//TODO close channels
	lg.Rabbit.Stop()
	lib.SugarLogger.Infof("LoginGate %d Closed.", lg.Id)
}

func (lg *LoginGate) Start() (err error) {
	lg.ServerCommon.Start()
	if lg.Rabbit == nil {
		lg.Rabbit = lib.NewRabbitClient()
	}
	err = lg.Rabbit.Start(config.RabbitUrl, "exchange", "queue", "fanout")
	lib.FatalOnError(err, "Start RabbitMQ Error")
	//conn, err := net.Dial("tcp", config.GameServerAddr+config.GameServerPort)
	//lib.FatalOnError(err, "logingate connect to gameserver error")
	//if conn != nil {
	//	conn.Write([]byte("Hello GameServer."))
	//}
	//defer conn.Close()
	lg.Run()
	return err
}

func (lg *LoginGate) Run() {
	for {
		select {
		case <-lg.SvrTick.C:
			go func() {
				ch, err := lg.Rabbit.SimpleConsume("queue", "")
				if ch != nil && err == nil {
					msg := <-ch
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
						lib.Log(zapcore.DebugLevel, fmt.Sprintln(p), err)
					case "protobuf":
						var p1 pb.Person
						if err := proto.Unmarshal(msg.Body, &p1); err != nil {
							lib.Log(zapcore.DebugLevel, "proto unmarshal msg error", err)
						}
						lib.Log(zapcore.DebugLevel, fmt.Sprintln(p1), err)
					default:
						lib.Log(zap.DebugLevel, "nothing", nil)
					}
				}
			}()
		case <-lg.CloseChan:
			lg.Stop()
		}
	}
}

func (lg *LoginGate) LoadConfig(path string) error {
	return nil
}
