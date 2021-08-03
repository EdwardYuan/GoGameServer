package global

import "GoGameServer/src/lib"

const ProjectName = "GoGameServer"

var serviceString map[string]ServerType
var ServerMap *ServerMapAddress

func makeSvcStringMap() {
	serviceString = make(map[string]ServerType)
	serviceString["game"] = ServerGame
	serviceString["login"] = ServerLogin
	serviceString["dbserver"] = ServerDatabase
	serviceString["gate"] = ServerGate

}

func Init(svrName string) {
	lib.InitLogger(svrName)
	makeSvcStringMap()
	ServerMap = NewServerMapAddress()
}
