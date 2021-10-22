package service_proxy

import (
	"GoGameServer/src/lib"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/coreos/etcd/client"
)

// 代理服务，主要用于服务注册与发现, 消息分发
type ServiceProxy struct {
	ProcessId int            // 进程ID ， 单机调试时用来标志每一个服务
	info      Serverinfo     // 服务端信息
	KeysAPI   client.KeysAPI // API client, 此处用的是 V2 版本的API，是基于 http 的。 V3版本的是基于grpc的API
}

// workerInfo is the service register information to etcd
type Serverinfo struct {
	Id   int32  `json:"id"`   // 服务器ID
	IP   string `json:"ip"`   // 对外连接服务的 IP
	Port int32  `json:"port"` // 对外服务端口，本机或者端口映射后得到的
}

// 注册服务
func RegisterService(endpoints []string) {
	cfg := client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		lib.FatalOnError(err, "Error: cannot connec to etcd:")
	}

	s := &ServiceProxy{
		ProcessId: os.Getpid(),
		info:      Serverinfo{Id: 1024, IP: "127.0.0.1", Port: 100},
		KeysAPI:   client.NewKeysAPI(etcdClient),
	}
	go s.HeartBeat() // 定时发送心跳
}

func (s *ServiceProxy) HeartBeat() {
	api := s.KeysAPI
	for {
		key := "lc_server/p_" + strconv.Itoa(s.ProcessId) // 先用 pid 来标识每一个服务， 通常应该用 IP 等来标识。
		// etcd 之所以适合用来做服务发现，是因为它是带目录结构的。 注册一类服务，
		// 只需要 key 在同一个目录下，此处 lc_sercer 目录下，p_{pid}
		value, _ := json.Marshal(s.info)

		_, err := api.Set(context.Background(), key, string(value), &client.SetOptions{
			TTL: time.Second * 20,
		}) // 调用 API， 设置该 key TTL 为20秒。

		if err != nil {
			lib.SugarLogger.Errorf("Error update workerInfo: %v", err)
		}
		time.Sleep(time.Second * 10)
	}
}
