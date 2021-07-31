package run

import (
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
	"GoGameServer/src/service_db"
	"GoGameServer/src/service_gate"
	"GoGameServer/src/service_gs"
	"GoGameServer/src/service_lg"
	"errors"
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
	serviceType := strings.ToLower(strings.TrimSpace(os.Args[2]))
	serviceIdx, err := strconv.Atoi(os.Args[3])
	if err != nil {
		serviceIdx = 1
	}
	serviceName := fmt.Sprintf(serviceType+"_%d", serviceIdx)
	// Init Global Variables
	var Svr service_common.Service
	global.Init()
	addr := lib.GetLocalIP(lib.IPv4)
	if addr == "" {
		lib.FatalOnError(errors.New(""), "get local ip address error")
	}
	lib.SugarLogger.Info("IP address: " + addr)
	global.ServerMap.MapAddrToServerName(lib.GetLocalIP(lib.IPv4), serviceType, serviceName)

	switch serviceType {
	case "game":
		Svr = service_gs.NewGameServer(serviceName, serviceIdx)
	case "login":
		Svr = service_lg.NewLoginGate(serviceName, serviceIdx)
	case "dbserver":
		Svr = service_db.NewServiceDB(serviceName, serviceIdx)
	case "gate":
		Svr = service_gate.NewServiceGate(serviceName, serviceIdx)
	default:
		fmt.Printf("GoGameServer: parameter error\n")
		return
	}
	lib.FatalOnError(Svr.Start(), "Start service error")
	fmt.Printf("%s service %s start...", global.ProjectName, serviceType)

}
