package main

import (
	"GoGameServer/src/service_gs"
	"fmt"
)

const ProjectName = "Common Game"

func main() {
	fmt.Printf("%s service start...", ProjectName)
	gs := service_gs.NewGameServer(fmt.Sprintf(ProjectName+"_%d", 1))
	gs.Start()
}
