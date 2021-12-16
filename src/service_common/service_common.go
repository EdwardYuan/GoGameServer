package service_common

import (
	"GoGameServer/src/config"
	"GoGameServer/src/lib"
	"time"

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
	SvrTick   *time.Ticker
}

func (s *ServerCommon) Stop() {
}

func (s *ServerCommon) LoadConfig(path string) error {
	viper.AddConfigPath(".")
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	config.RabbitUrl = viper.GetString("rabbitmq.url")
	config.GameServerAddr = viper.GetString("gameserver.addr")
	config.GameServerPort = viper.GetString("gameserver.port")
	config.LoginGateAddr = viper.GetString("logingate.addr")
	config.LoginGatePort = viper.GetString("logingate.port")
	lib.FatalOnError(err, "Load Config Error")
	return nil
}

func (s *ServerCommon) Start() (err error) {
	lib.FatalOnError(s.LoadConfig("./config"), "Load Config Error")
	s.SvrTick = time.NewTicker(time.Duration(time.Millisecond))
	return
}

func (s *ServerCommon) Run() {

}
