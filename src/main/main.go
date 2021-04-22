package main

import (
	"GoGameServer/src/service_gs"
	"fmt"
)

const PROJECT_NAME = "Common Game"

func main() {
	fmt.Printf("%s service start...", PROJECT_NAME)
	gs := service_gs.NewGameServer(fmt.Sprintf(PROJECT_NAME+"%d", 1))
	gs.Start()
}
