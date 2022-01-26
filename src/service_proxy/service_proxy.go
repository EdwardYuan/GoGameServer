package service_proxy

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/config"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	client "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"time"
)

// 代理服务，主要用于服务注册与发现, 消息分发
type ServiceProxy struct {
	ProcessId int        // 进程ID ， 单机调试时用来标志每一个服务
	info      Serverinfo // 服务端信息
	Servers   map[string]*Serverinfo
	Agent     *EtcdAgent
	workPool  *ants.Pool
	gnet.EventHandler
	GameConnections map[string]gnet.Conn
	AgentsToGames   map[uint64]string
	MsgChan         chan pb.ProtoInternal
}

type EtcdAgent struct {
	Proxy         *ServiceProxy
	RegisteredSvr chan Serverinfo
	QueryChan     chan int32
	Client        *client.Client
	ticker        time.Ticker
	CloseChan     chan bool
}

func NewEtcdAgent() *EtcdAgent {
	cli, err := client.New(client.Config{
		Endpoints:   []string{"http://127.0.0.1:2359"}, // TODO 读取配置
		DialTimeout: 5 * time.Second,
	})
	lib.FatalOnError(err, "New Proxy Service error")
	return &EtcdAgent{
		Client:        cli,
		RegisteredSvr: make(chan Serverinfo, 100),
		ticker:        *time.NewTicker(20 * time.Second),
	}
}

func NewServiceProxy(_name string, id int) *ServiceProxy {
	pool, err := ants.NewPool(ants.DefaultAntsPoolSize)
	lib.FatalOnError(err, "Create Proxy Service error")
	return &ServiceProxy{
		ProcessId:       0, // 自己的ProcessId为0
		info:            NewServerInfo(int32(id), lib.GetLocalIP(lib.IPv4), _name, 0),
		Servers:         make(map[string]*Serverinfo, 1),
		Agent:           NewEtcdAgent(),
		workPool:        pool,
		MsgChan:         make(chan pb.ProtoInternal, lib.MaxMessageCount),
		GameConnections: make(map[string]gnet.Conn, lib.MaxGameServerCount),
		AgentsToGames:   make(map[uint64]string, lib.MaxTotalAgents),
	}
}

func (s *ServiceProxy) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	if s.GameConnections != nil {
		addr := c.RemoteAddr()
		serverInfo := s.Servers[addr.String()]
		s.GameConnections[serverInfo.Name] = c
	}
	return
}

func (s *ServiceProxy) SendToGame(name string, sessionId uint64, data []byte) {
	if conn, ok := s.GameConnections[name]; ok {
		if err := conn.AsyncWrite(data); err != nil { // 异步写会不会有问题，如果客户端发来的消息依赖顺序
			lib.LogErrorAndReturn(err, "ServiceProxy SendToGame Error")
		}
	}
}

func (c *EtcdAgent) GetServerInfo(name string) *Serverinfo {
	resp, err := c.Client.Get(context.TODO(), "services/"+name)
	var serverStr string

	lib.LogErrorAndReturn(err, "Etcd Agent GetServerInfo")
	for _, v := range resp.Kvs {
		json.Unmarshal(v.Value, &serverStr)
	}
	ServerInfo := makeServerInfo(serverStr)
	return ServerInfo
}

func makeServerInfo(value string) *Serverinfo {
	var err error
	infos := strings.Split(value, ",")
	id, err1 := strconv.Atoi(infos[0])
	port, err2 := strconv.Atoi(infos[3])
	if err1 != nil {
		err = err1
	}
	if err2 != nil {
		err = err2
	}
	lib.LogErrorAndReturn(err, "makeServerInfo")
	return &Serverinfo{
		Id:   int32(id),
		Name: infos[1],
		IP:   infos[2],
		Port: int32(port),
	}
}

func buildServerInfoString(s *Serverinfo) string {
	return strconv.Itoa(int(s.Id)) + "," + s.Name + "," + s.IP + strconv.Itoa(int(s.Port))
}

func (c *EtcdAgent) run(s *Serverinfo) {
	for {
		select {
		case <-c.RegisteredSvr:
			go func() {
				_, err := c.Client.Put(context.TODO(), "services/"+s.Name, buildServerInfoString(s))
				lib.LogIfError(err, "Register server error")
			}()
		case serverId := <-c.QueryChan:
			go func() {
				key := "services/" + strconv.Itoa(int(serverId))
				resp, err := c.Client.Get(context.Background(), key)
				lib.LogIfError(err, "server not registered")
				// TODO 查询结果返回 处理相应的数据 到这里说明查到了注册的服务
				value := resp.Kvs[0]
				// TODO parse value and return
				fmt.Printf("serverinfo is %+v", value)
			}()
		case <-c.ticker.C:
			go c.Proxy.HeartBeat()
		case <-c.CloseChan:
			c.Client.Close()
		}
	}
}

func (p *ServiceProxy) Start() (err error) {
	p.Agent.Proxy = p
	p.AddrServer(&p.info) // 首先添加自身服务到etcd
	go func() {
		if gnet.Serve(p, config.ProxyAddr, gnet.WithCodec(codec.CodecProtobuf{}),
			gnet.WithMulticore(true)) != nil {
			lib.FatalOnError(err, "Proxy Serve error")
		}
	}()
	defer func() {
		p.workPool.Release()
	}()
	p.Run()
	return
}

func (p *ServiceProxy) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	go p.workPool.Submit(
		func() {
			var message *pb.ProtoInternal
			if lib.LogErrorAndReturn(proto.Unmarshal(frame, message), "unmarshal message error") {
				return
			}
			switch message.Cmd {
			case pb.InternalGateToProxy:
				//  登录游戏 第一次 创建连接 加入map
				if p.findGameServerBySession(message.SessionId) == "" {
					err := p.RegisterSessionToGame(message.SessionId, message.Dst)
					if lib.LogErrorAndReturn(err, "register session to game") {
						return
					}
				}
				dst := message.Dst
				if service, ok := p.Servers[dst]; ok {
					if strings.Contains(service.Name, "game") {
						postMsg := pb.ProtoInternal{
							Cmd:       pb.InternalProxyToGame,
							Dst:       dst,
							SessionId: message.SessionId,
							Data:      frame,
						}
						p.MsgChan <- postMsg
					}
				}
			}
		})
	return
}

func (p *ServiceProxy) findGameServerBySession(sessionId uint64) (gameName string) {
	return p.AgentsToGames[sessionId]
}

func (p *ServiceProxy) RegisterSessionToGame(session uint64, name string) error {
	if p.AgentsToGames[session] != "" {
		return errors.New("already exists")
	}
	p.AgentsToGames[session] = name
	return nil
}

func (p *ServiceProxy) Stop() {

}

func (p *ServiceProxy) Run() {
	go p.Agent.run(&p.info)
	for {
		select {
		case msg := <-p.MsgChan:
			p.SendToGame(msg.Dst, msg.SessionId, msg.Data)
		}
	}
}

func (p *ServiceProxy) LoadConfig(path string) error {
	return nil
}

func (p *ServiceProxy) AddrServer(s *Serverinfo) {
	// 如果proxy服务的etcd client不存在，直接退出
	if p.Agent == nil {
		err := errors.New("no etcd agent exist")
		lib.FatalOnError(err, "Register new service")
	}
	key := strconv.Itoa(int(s.Id)) + s.Name
	// 注册对应的服务到
	p.Agent.RegisteredSvr <- *s
	p.Servers[key] = s
}

func (s *ServiceProxy) HeartBeat() {
	key := "services/" + s.info.Name
	if s.Agent != nil && s.Agent.Client != nil {
		_, err := s.Agent.Client.Get(context.Background(), key, nil)
		lib.FatalOnError(err, "Proxy Service"+strconv.Itoa(s.ProcessId)+" error: ")
	}
	var value string
	var leaseId client.LeaseID
	_, err := s.Agent.Client.Put(context.Background(), key, string(value), client.WithLease(leaseId))
	lib.FatalOnError(err, "Error update workerInfo: %v")
}

// ServerInfo is the service register information to etcd
type Serverinfo struct {
	Id   int32  `json:"id"`   // 服务器ID
	Name string `json:"name"` // 服务名
	IP   string `json:"ip"`   // 对外连接服务的 IP
	Port int32  `json:"port"` // 对外服务端口，本机或者端口映射后得到的
}

func NewServerInfo(id int32, ip string, name string, port int32) Serverinfo {
	return Serverinfo{
		Id:   id,
		Name: name,
		IP:   ip,
		Port: port,
	}
}
