package game

import (
	"GoGameServer/src/lib"
	"github.com/spf13/viper"
	"os"
)

type GSServer struct {
	Agents map[int64]*Agent
	MapEng *MapEngine
}

func NewGsServer() *GSServer {
	return &GSServer{
		Agents: make(map[int64]*Agent),
		MapEng: NewMapEngine(),
	}
}

func (gs *GSServer) LoadMaps() error {
	return nil
}

// LoadConfigs 这里加载的是游戏业务相关的配置，与service_game的config不同
func (gs *GSServer) LoadConfigs(path string) error {
	file, err := os.Open(path)
	lib.LogIfError(err, "load gs config error")
	viper.AddConfigPath(path)
	err = viper.ReadConfig(file)
	return err
}
