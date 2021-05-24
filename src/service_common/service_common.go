package service_common

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/config"
	"GoGameServer/src/lib"
	"github.com/panjf2000/gnet"
	"github.com/spf13/viper"
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

func (s *ServerCommon) Start() {
	s.LoadConfig("./config")
	s.Rabbit = lib.NewRabbitClient()
	s.Rabbit.Start(config.RabbitUrl, "exchange", "queue", "fanout")
}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

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
