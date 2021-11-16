package service_proxy

import (
	"GoGameServer/src/lib"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	client "go.etcd.io/etcd/client/v3"
)

// 代理服务，主要用于服务注册与发现, 消息分发
type ServiceProxy struct {
	ProcessId int        // 进程ID ， 单机调试时用来标志每一个服务
	info      Serverinfo // 服务端信息
	Servers   map[string]*Serverinfo
	Agent     *EtcdAgent
}

type EtcdAgent struct {
	Proxy         *ServiceProxy
	RegisteredSvr chan Serverinfo
	QueryChan     chan int32
	Client        *client.Client
	ticker        time.Ticker
}

func NewEtcdAgent() *EtcdAgent {
	cli, err := client.New(client.Config{
		Endpoints:   []string{"http://127.0.0.1:2359"},
		DialTimeout: 5 * time.Second,
	})
	lib.FatalOnError(err, "New Proxy Service error")
	return &EtcdAgent{
		Client:        cli,
		RegisteredSvr: make(chan Serverinfo, 100),
		ticker:        *time.NewTicker(20 * time.Second),
	}
}

func NewSericeProxy(_name string, id int) *ServiceProxy {
	return &ServiceProxy{
		ProcessId: 0, // 自己的processid为0
		info:      NewSericeInfo(int32(id), lib.GetLocalIP(lib.IPv4), _name, 0),
		Servers:   make(map[string]*Serverinfo, 1),
		Agent:     NewEtcdAgent(),
	}
}

func (c *EtcdAgent) run(s *Serverinfo) {
	for {
		select {
		case <-c.RegisteredSvr:
			go func() {
				_, err := c.Client.Put(context.TODO(), "services/"+strconv.Itoa(int(s.Id)), s.IP+":"+strconv.Itoa(int(s.Port)))
				if err != nil {
					lib.SugarLogger.Errorf("Register server error %v", err)
				}
			}()
		case serverId := <-c.QueryChan:
			go func() {
				key := "services/" + strconv.Itoa(int(serverId))
				resp, err := c.Client.Get(context.Background(), key)
				if err != nil {
					lib.SugarLogger.Errorf("server not registered %v", err)
				}
				// TODO 查询结果返回 处理相应的数据 到这里说明查到了注册的服务
				value := resp.Kvs[0]
				// TODO parse value and return
				fmt.Printf("serverinfo is %+v", value)
			}()
		case <-c.ticker.C:
			go c.Proxy.HeartBeat()
		}
	}
}

func (p *ServiceProxy) Start() (err error) {
	p.Agent.Proxy = p
	p.AddrServer(&p.info) // 首先添加自身服务到etcd
	p.Run()
	return
}

func (p *ServiceProxy) Stop() {

}

func (p *ServiceProxy) Run() {
	p.Agent.run(&p.info)
}

func (p *ServiceProxy) LoadConfig(path string) error {
	return nil
}

func (p *ServiceProxy) AddrServer(s *Serverinfo) {
	// 如果proxy服务的etcd client不存在，直接退出
	if p.Agent == nil {
		err := errors.New("No EtcdAgent Exist.")
		lib.FatalOnError(err, "Register new service")
	}
	key := strconv.Itoa(int(s.Id)) + s.Name
	// 注册对应的服务到
	p.Agent.RegisteredSvr <- *s
	p.Servers[key] = s
}

func (s *ServiceProxy) HeartBeat() {
	key := "services/" + strconv.Itoa(s.ProcessId)
	_, err := s.Agent.Client.Get(context.Background(), key, nil)
	lib.FatalOnError(err, "Proxy Service"+strconv.Itoa(s.ProcessId)+" error: ")
	var value string
	var leaseId client.LeaseID
	_, err = s.Agent.Client.Put(context.Background(), key, string(value), client.WithLease(leaseId))
	lib.FatalOnError(err, "Error update workerInfo: %v")
}

// ServerInfo is the service register information to etcd
type Serverinfo struct {
	Id   int32  `json:"id"`   // 服务器ID
	Name string `json:"name"` // 服务名
	IP   string `json:"ip"`   // 对外连接服务的 IP
	Port int32  `json:"port"` // 对外服务端口，本机或者端口映射后得到的
}

func NewSericeInfo(id int32, ip string, name string, port int32) Serverinfo {
	return Serverinfo{
		Id:   id,
		Name: name,
		IP:   ip,
		Port: port,
	}
}
