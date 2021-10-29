package service_proxy

import (
	"GoGameServer/src/lib"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	client "go.etcd.io/etcd/clientv3"
)

// 代理服务，主要用于服务注册与发现, 消息分发
type ServiceProxy struct {
	ProcessId int        // 进程ID ， 单机调试时用来标志每一个服务
	info      Serverinfo // 服务端信息
	// KeysAPI   client.KeysAPI // API client, 此处用的是 V2 版本的API，是基于 http 的。 V3版本的是基于grpc的API
	Servers map[int32]*Serverinfo
	Center  *RegisterCenter
}

type RegisterCenter struct {
	Proxy         *ServiceProxy
	RegisteredSvr chan Serverinfo
	QueryChan     chan int32
	Client        *client.Client
}

func NewRegisterCenter() *RegisterCenter {
	cli, err := client.New(client.Config{
		Endpoints:   []string{"http://127.0.0.1:2359"},
		DialTimeout: 5 * time.Second,
	})
	lib.FatalOnError(err, "New Proxy Service error")
	return &RegisterCenter{
		Client:        cli,
		RegisteredSvr: make(chan Serverinfo, 100),
	}
}

func NewSericeProxy(_name string, id int) *ServiceProxy {
	return &ServiceProxy{
		ProcessId: 0,
		info:      NewSericeInfo(0, "", 0),
		Servers:   make(map[int32]*Serverinfo, 1),
		Center:    NewRegisterCenter(),
	}
}

func (c *RegisterCenter) run(s *Serverinfo) {
	for {
		select {
		case <-c.RegisteredSvr:
			go func() {
				_, err := c.Client.Put(context.TODO(), "services"+strconv.Itoa(int(s.Id)), s.IP+":"+strconv.Itoa(int(s.Port)))
				if err != nil {
					lib.SugarLogger.Errorf("Register server error %v", err)
				}
			}()
		case serverId := <-c.QueryChan:
			go func() {
				key := "services" + strconv.Itoa(int(serverId))
				resp, err := c.Client.Get(context.Background(), key)
				if err != nil {
					lib.SugarLogger.Errorf("server not registered %v", err)
				}
				// TODO 查询结果返回 处理相应的数据 到这里说明查到了注册的服务
				value := resp.Kvs[0]
				// TODO parse value and return
				fmt.Printf("serverinfo is %+v", value)
			}()
		}
	}
}

func (p *ServiceProxy) Start() (err error) {
	p.Center.Proxy = p
	p.Run()
	return
}

func (p *ServiceProxy) Stop() {

}

func (p *ServiceProxy) Run() {
	p.Center.run(&p.info)
}

func (p *ServiceProxy) LoadConfig(path string) error {
	return nil
}

func (p *ServiceProxy) AddrServer(s *Serverinfo) {
	// 如果proxy服务的etcd client不存在，直接退出
	if p.Center == nil {
		err := errors.New("No RegisterCenter Exist.")
		lib.FatalOnError(err, "Register new service")
	}
	// 注册对应的服务到
	p.Servers[s.Id] = s
	p.Center.RegisteredSvr <- *s
}

// workerInfo is the service register information to etcd
type Serverinfo struct {
	Id   int32  `json:"id"`   // 服务器ID
	IP   string `json:"ip"`   // 对外连接服务的 IP
	Port int32  `json:"port"` // 对外服务端口，本机或者端口映射后得到的
}

func NewSericeInfo(id int32, ip string, port int32) Serverinfo {
	return Serverinfo{
		Id:   id,
		IP:   ip,
		Port: port,
	}
}

/*
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
*/
