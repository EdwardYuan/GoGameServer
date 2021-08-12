package service_common

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	SvrTick   *time.Ticker
	*gnet.EventServer
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

func (s *ServerCommon) Start() {
	lib.FatalOnError(s.LoadConfig("./config"), "Load Config Error")
	s.SvrTick = time.NewTicker(time.Duration(time.Millisecond))
}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	msg, err := s.Decode(frame)   // 从[]byte解析出Message消息，之后分发给相应的服务
	lib.Log(zapcore.DebugLevel, "gnet receive message", err)
	fmt.Println(msg) // to remove
	switch global.ServerMap.GetSvrTypeByAddr(c.RemoteAddr().String()) {
	case global.ServerDatabase:
	case global.ServerGame:

	case global.ServerGate:

	case global.ServerLogin:
	default:

	}
	return
}

func (s *ServerCommon) Encode(msg MsgHandler.Message) (data []byte, err error) {
	return
}

func (s *ServerCommon) Decode(data []byte) (msg MsgHandler.Message, err error) {
	head := MsgHandler.NewMessageHead()
	head.Decode(data)
	err = head.Check()

	lib.Log(zap.ErrorLevel, "Decode Message Data Error: ", err)
	return
}

func (s *ServerCommon) HandleMessage(msg MsgHandler.Message) {

}
