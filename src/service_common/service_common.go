package service_common

import (
	"GoGameServer/src/config"
	"strings"
	"sync"
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
	closeOnce sync.Once
}

// NewServerCommon 创建并初始化一个 ServerCommon 实例
// 参数:
//
//	name - 服务器名称
//	id - 服务器ID
//
// 返回值:
//
//	*ServerCommon - 初始化后的 ServerCommon 指针
func NewServerCommon(name string, id int) *ServerCommon {
	return &ServerCommon{
		Name:      name,
		Id:        id,
		CloseChan: make(chan int, 1),
	}
}

// InitCommon 初始化服务器通用组件
// 该方法会初始化关闭通道、加载配置文件以及服务器时钟
// 参数:
//
//	configPath - 配置文件路径
//
// 返回值:
//
//	error - 如果加载配置失败则返回错误，否则返回nil
func (s *ServerCommon) InitCommon(configPath string) error {
	// 检查并初始化关闭通道
	if s.CloseChan == nil {
		s.CloseChan = make(chan int, 1)
	}
	// 加载配置文件
	if err := s.LoadConfig(configPath); err != nil {
		return err
	}
	// 检查并初始化服务器时钟
	if s.SvrTick == nil {
		s.SvrTick = time.NewTicker(time.Millisecond)
	}
	return nil
}

// Stop 停止服务器通用组件
// 该方法确保服务器的ticker被停止，并且只执行一次停止操作
// 使用sync.Once保证即使多次调用Stop方法也只会执行一次实际的停止逻辑
func (s *ServerCommon) Stop() {
	s.closeOnce.Do(func() {
		if s.SvrTick != nil {
			s.SvrTick.Stop()
		}
	})
}

func (s *ServerCommon) NotifyStop() {
	if s.CloseChan == nil {
		return
	}
	// 使用非阻塞发送，避免重复停止或无人接收时卡住调用方。
	select {
	case s.CloseChan <- 1:
	default:
	}
}

// LoadConfig 只负责把配置读入全局 config，不在这里 fatal。
// 各服务的 Start 根据返回值决定是否启动失败。
func (s *ServerCommon) LoadConfig(path string) error {
	viper.AddConfigPath(".")
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	config.RabbitUrl = viper.GetString("rabbitmq.url")
	config.GameServerAddr = viper.GetString("gameserver.addr")
	config.GameServerPort = viper.GetString("gameserver.port")
	config.LoginGateAddr = viper.GetString("logingate.addr")
	config.LoginGatePort = viper.GetString("logingate.port")
	config.GameGateAddr = viper.GetString("gamegate.addr")
	config.GameGatePort = viper.GetString("gamegate.port")
	config.DBServiceAddr = viper.GetString("dbserver.addr")
	config.DBServicePort = viper.GetString("dbserver.port")
	config.MySqlUrl = viper.GetString("mysql.addr")
	config.EtcdUrl = viper.GetString("etcd.endpoints")
	config.ProxyAddr = viper.GetString("proxy.addr")
	config.ProxyPort = viper.GetString("proxy.port")
	// 兼容当前配置里 proxy.addr 已经包含端口的写法。
	if config.ProxyPort == "" && strings.Contains(config.ProxyAddr, ":") {
		parts := strings.Split(config.ProxyAddr, ":")
		config.ProxyPort = parts[len(parts)-1]
	}
	return nil
}

func (s *ServerCommon) Start() error {
	return s.InitCommon("./config")
}

//func (s *ServerCommon) Encode(msg lib.Message) (data []byte, err error) {
//	return
//}
//
//func (s *ServerCommon) Decode(data []byte) (msg lib.Message, err error) {
//	head := lib.NewMessageHead()
//	head.Decode(data)
//	err = head.Check()
//
//	lib.Log(zap.ErrorLevel, "Decode Message Data Error: ", err)
//	return
//}
//
//func (s *ServerCommon) HandleMessage(msg lib.Message) {
//	go func() {
//
//	}()
//}
