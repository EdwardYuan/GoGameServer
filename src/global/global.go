package global

import (
	"GoGameServer/src/lib"
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

const ProjectName = "GoGameServer"
const DefaultPoolSize = 1024

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

func RegisterService(ep []string, name string, typ string, desc string) {
	if lib.Logger == nil {
		lib.FatalOnError(errors.New("logger not initialized"), "logger not assigned")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   ep,
		DialTimeout: 5 * time.Second,
		Logger:      lib.Logger,
	})
	lib.FatalOnError(err, "failed to create etcd client.")
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {
			lib.LogIfError(err, "Close client error")
		}
	}(cli)
	_, err = cli.Put(context.Background(), name, typ)
	lib.LogIfError(err, name+"failed to register service to etcd")
}
