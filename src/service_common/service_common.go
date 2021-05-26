package service_common

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"time"
)

type Service interface {
	Start() (err error)
	Stop()
	Run()
	LoadConfig(path string) error
}

type ServerCommon struct {
	Name      string
	Id        int
	CloseChan chan int
	Rabbit    *lib.RabbitClient
	SvrTick   *time.Ticker
	*gnet.EventServer
}

func (s *ServerCommon) Stop() {
	s.Rabbit.Stop()
}

func (s *ServerCommon) LoadConfig(path string) error {
	viper.AddConfigPath(".")
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	config.RabbitUrl = viper.GetString("rabbitmq.url")
	lib.FatalOnError(err, "Load Config Error")
	return nil
}

func (s *ServerCommon) Run() {
	select {
	case <-s.SvrTick.C:
		s.Rabbit.SimpleConsume("queue", "")
	case <-s.CloseChan:
		s.Stop()
	}
}

func (s *ServerCommon) Start() {
	s.LoadConfig("./config")
	s.SvrTick = time.NewTicker(time.Duration(time.Millisecond))
	s.Rabbit = lib.NewRabbitClient()
	s.Rabbit.Start(config.RabbitUrl, "exchange", "queue", "fanout")
}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg, err := s.Decode(frame)
	lib.Log(zapcore.DebugLevel, "gnet receive message", err)
	fmt.Println(msg) // to remove
	switch global.ServerMap.GetSvrTypeByAddr(c.RemoteAddr()) {
	case global.ServerDatabase:
	case global.ServerGame:
	case global.ServerGate:

	case global.ServerLogin:
	default:

	}
	return
}

func (s *ServerCommon) Encode(msg MsgHandler.Message) (err error) {
	//TODO protobuf marshal
	return
}

func (s *ServerCommon) Decode(data []byte) (msg MsgHandler.Message, err error) {
	//TODO protobuf unmarshal
	return
}

func (s *ServerCommon) HandleMessage(msg MsgHandler.Message) {

}
