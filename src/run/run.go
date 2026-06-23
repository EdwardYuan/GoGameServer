package run

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
	"GoGameServer/src/service_db"
	"GoGameServer/src/service_game"
	"GoGameServer/src/service_gate"
	"GoGameServer/src/service_login"
	"GoGameServer/src/service_proxy"
	"go.uber.org/zap"
)

// StartServer 启动指定类型的游戏服务器服务
// 参数:
//   args []string - 命令行参数，至少需要4个参数：程序名、端口、服务类型、服务索引
//
// 该函数根据命令行参数启动不同类型的服务（如game、login、dbserver、gate、proxy）
// 并处理服务初始化和启动过程中的错误
func StartServer(args []string) {
	if len(args) < 4 {
		log.Fatal("not enough parameters, please specify the service to start.")
		return
	}
	// 解析服务类型和服务索引
	serviceType := strings.ToLower(strings.TrimSpace(args[2]))
	serviceIdx, err := strconv.Atoi(args[3])
	if err != nil {
		serviceIdx = 1
	}
	serviceName := fmt.Sprintf(serviceType+"_%d", serviceIdx)
	// Init Global Variables
	var Svr service_common.Service
	// 初始化全局变量
	global.Init(serviceName)
	// 获取本地IP地址
	addr := lib.GetLocalIP(lib.IPv4)
	if addr == "" {
		lib.SysLoggerFatal(err, "get local ip address error")
	}
	lib.SugarLogger.Info("IP address: " + addr)
	// 将IP地址映射到服务器名称
	global.ServerMap.MapAddrToServerName(lib.GetLocalIP(lib.IPv4), serviceType, serviceName)
	// 记录服务启动日志
	lib.Log(zap.InfoLevel, "starting "+serviceType, nil)
	// 根据服务类型创建对应的服务实例
	switch serviceType {
	case "game":
		Svr = service_game.NewGameServer(serviceName, serviceIdx)
	case "login":
		Svr = service_login.NewLoginGate(serviceName, serviceIdx)
	case "dbserver":
		Svr = service_db.NewServiceDB(serviceName, serviceIdx)
	case "gate":
		Svr = service_gate.NewServiceGate(serviceName, serviceIdx)
	case "proxy":
		Svr = service_proxy.NewServiceProxy(serviceName, serviceIdx)
	default:
		fmt.Printf("GoGameServer: parameter error\n")
		return
	}
	// 启动服务并检查错误
	lib.SysLoggerFatal(Svr.Start(), "Start service error")
	fmt.Printf("%s service %s start...", global.ProjectName, serviceType)

}
