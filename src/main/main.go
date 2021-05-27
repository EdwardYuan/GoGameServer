package main

import (
	"GoGameServer/src/global"
	"GoGameServer/src/service_gs"
	"GoGameServer/src/service_lg"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const ProjectName = "GoGameServer"

func main() {
	if len(os.Args) < 1 {
		log.Fatal("not enough parameters, please specify the service to start.")
		return
	}
	serviceName := os.Args[1]
	serviceIdx, err := strconv.Atoi(os.Args[2])
	if err != nil {
		serviceIdx = 1
	}
	// Init Global Variables
	global.GlobalInit()
	switch strings.ToLower(serviceName) {
	case "game":
		gs := service_gs.NewGameServer(fmt.Sprintf(serviceName+"_%d", serviceIdx), serviceIdx)
		gs.Start()
	case "login":
		lg := service_lg.NewLoginGate(fmt.Sprintf(serviceName+"_%d", serviceIdx), serviceIdx)
		lg.Start()

	}
	fmt.Printf("%s service %s start...", ProjectName, serviceName)
}
