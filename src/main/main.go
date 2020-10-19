package main

import (
	"fmt"
	"github.com/panjf2000/gnet"
)

const PROJECT_NAME = "Common Game"

type GameServer struct {
	*gnet.EventServer
}

func main() {
	fmt.Printf("%s service start...", PROJECT_NAME)
	gs := &GameServer{}
	gnet.Serve(gs, "tcp://127.0.0.1:9000", gnet.WithMulticore(true))

}
