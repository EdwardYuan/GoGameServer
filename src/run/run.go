package run

import (
	"GoGameServer/src/global"
	"GoGameServer/src/service_common"
	"GoGameServer/src/service_db"
	"GoGameServer/src/service_gs"
	"GoGameServer/src/service_lg"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func RunServer(args []string) {
	if len(args) < 3 {
		log.Fatal("not enough parameters, please specify the service to start.")
		return
	}
	serviceName := strings.TrimSpace(os.Args[2])
	serviceIdx, err := strconv.Atoi(os.Args[3])
	if err != nil {
		serviceIdx = 1
	}
	// Init Global Variables
	var Svr service_common.Service
	global.GlobalInit()
	switch strings.ToLower(serviceName) {
	case "game":
		Svr = service_gs.NewGameServer(fmt.Sprintf(serviceName+"_%d", serviceIdx), serviceIdx)
	case "login":
		Svr = service_lg.NewLoginGate(fmt.Sprintf(serviceName+"_%d", serviceIdx), serviceIdx)
	case "dbserver":
		Svr = service_db.NewServiceDB(fmt.Sprintf(serviceName+"_%d", serviceIdx), serviceIdx)
	default:
		fmt.Printf("GoGameServer: parameter error\n")
		return
	}
	Svr.Start()
	fmt.Printf("%s service %s start...", global.ProjectName, serviceName)

}
