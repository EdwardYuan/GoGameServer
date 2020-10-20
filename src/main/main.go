package main

import (
	"fmt"
	"github.com/panjf2000/gnet"
	"gogameserver/service_common"
)

const PROJECT_NAME = "Common Game"

type GameServer struct {
	*service_common.ServerCommon
}

func NewGameServer(service *service_common.ServerCommon) *GameServer {
	return &GameServer{ServerCommon: service}
}

func main() {
	fmt.Printf("%s service start...", PROJECT_NAME)
	gs := NewGameServer(new(service_common.ServerCommon))
	gnet.Serve(gs, "tcp://127.0.0.1:9000", gnet.WithMulticore(true))

}
