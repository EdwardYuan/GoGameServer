package service_proxy

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/config"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"context"
	"errors"
	"github.com/panjf2000/ants/v2"
	gnet "github.com/panjf2000/gnet/v2"
	client "go.etcd.io/etcd/client/v3"
	"net"
	"strings"
	"sync"
	"time"
)

// 代理服务，主要用于服务注册与发现, 消息分发
type ServiceProxy struct {
	ProcessId int        // 进程ID ， 单机调试时用来标志每一个服务
	info      ServerInfo // 服务端信息
	*service_common.ServerCommon
	Servers  map[string]*ServerInfo
	Agent    *EtcdAgent
	workPool *ants.Pool
	gnet.EventHandler
	// ServiceConnections 保存已经通过 InternalProxySync 握手的服务连接。
	// key 使用 serviceName，例如 game_1、gate_1。
	ServiceConnections map[string]gnet.Conn
	MsgChan            chan pb.ProtoInternal
	mu                 sync.RWMutex
	stopOnce           sync.Once
}

type EtcdAgent struct {
	Proxy *ServiceProxy
	// RegisteredSvr 接收需要写入 etcd 的服务信息。
	RegisteredSvr chan ServerInfo
	// QueryChan 使用 serviceName 查询 services/{serviceName}，不再按 id 查询。
	QueryChan chan string
	Client    *client.Client
	ticker    time.Ticker
	CloseChan chan bool
}

func NewEtcdAgent() *EtcdAgent {
	return &EtcdAgent{
		RegisteredSvr: make(chan ServerInfo, 100),
		QueryChan:     make(chan string, 100),
		ticker:        *time.NewTicker(20 * time.Second),
		CloseChan:     make(chan bool, 1),
	}
}

func (c *EtcdAgent) Start(endpoint string) error {
	if endpoint == "" {
		endpoint = "127.0.0.1:2379"
	}
	cli, err := client.New(client.Config{
		Endpoints:   []string{service_common.NormalizeEtcdEndpoint(endpoint)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	c.Client = cli
	return nil
}

func NewServiceProxy(_name string, id int) *ServiceProxy {
	pool, err := ants.NewPool(ants.DefaultAntsPoolSize)
	lib.FatalOnError(err, "Create Proxy Service error")
	return &ServiceProxy{
		ProcessId:          0, // 自己的ProcessId为0
		info:               NewServerInfo(int32(id), lib.GetLocalIP(lib.IPv4), _name, "proxy", 0),
		ServerCommon:       service_common.NewServerCommon(_name, id),
		Servers:            make(map[string]*ServerInfo, 1),
		Agent:              NewEtcdAgent(),
		workPool:           pool,
		MsgChan:            make(chan pb.ProtoInternal, lib.MaxMessageCount),
		ServiceConnections: make(map[string]gnet.Conn, lib.MaxGameServerCount),
	}
}

func (s *ServiceProxy) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return
}

func (s *ServiceProxy) SendToGame(name string, sessionId uint64, data []byte) {
	s.SendToService(name, pb.InternalProxyToGame, sessionId, data)
}

func (s *ServiceProxy) SendToGate(name string, sessionId uint64, data []byte) {
	s.SendToService(name, pb.InternalProxyToGate, sessionId, data)
}

func (s *ServiceProxy) SendToService(name string, cmd int32, sessionId uint64, data []byte) {
	// 转发必须依赖握手后的连接表，避免用 RemoteAddr 猜测服务身份。
	s.mu.RLock()
	conn, ok := s.ServiceConnections[name]
	s.mu.RUnlock()
	if ok {
		msg := &pb.ProtoInternal{
			Cmd:       cmd,
			Dst:       name,
			SessionId: sessionId,
			Data:      data,
		}
		packet, err := codec.EncodeMessage(msg)
		if err != nil {
			lib.LogErrorAndReturn(err, "ServiceProxy encode message error")
			return
		}
		if err := conn.AsyncWrite(packet, nil); err != nil {
			// 异步写会不会有问题，如果客户端发来的消息依赖顺序
			lib.LogErrorAndReturn(err, "ServiceProxy SendToService Error")
		}
		return
	}
	lib.SugarLogger.Warnf("ServiceProxy connection not found: %s", name)
}

func (c *EtcdAgent) GetServerInfo(name string) *ServerInfo {
	if c.Client == nil {
		return nil
	}
	resp, err := c.Client.Get(context.TODO(), service_common.ServiceKey(name))
	if lib.LogErrorAndReturn(err, "Etcd Agent GetServerInfo") {
		return nil
	}
	if len(resp.Kvs) == 0 {
		return nil
	}
	serverInfo := &ServerInfo{}
	for _, v := range resp.Kvs {
		decoded, err := service_common.UnmarshalServerInfo(string(v.Value))
		if err != nil {
			lib.LogIfError(err, "Etcd Agent decode ServerInfo")
			return nil
		}
		*serverInfo = *decoded
	}
	return serverInfo
}

func makeServerInfo(value string) *ServerInfo {
	if info, err := service_common.UnmarshalServerInfo(value); err == nil {
		return info
	}
	return nil
}

func buildServerInfoString(s *ServerInfo) (string, error) {
	return service_common.MarshalServerInfo(s)
}

func (c *EtcdAgent) run(s *ServerInfo) {
	for {
		select {
		case registered := <-c.RegisteredSvr:
			go func() {
				if c.Client == nil {
					return
				}
				value, err := buildServerInfoString(&registered)
				if err != nil {
					lib.LogIfError(err, "encode server info error")
					return
				}
				_, err = c.Client.Put(context.TODO(), service_common.ServiceKey(registered.Name), value)
				lib.LogIfError(err, "Register server error")
			}()
		case serviceName := <-c.QueryChan:
			go func() {
				// 查询路径统一为 services/{serviceName}，和注册路径保持一致。
				key := service_common.ServiceKey(serviceName)
				if c.Client == nil {
					return
				}
				resp, err := c.Client.Get(context.Background(), key)
				lib.LogIfError(err, "server not registered")
				if err != nil || len(resp.Kvs) == 0 {
					return
				}
				// TODO 查询结果返回 处理相应的数据 到这里说明查到了注册的服务
				value := resp.Kvs[0]
				// TODO parse value and return
				lib.SugarLogger.Debugf("serverinfo is %+v", value)
			}()
		case <-c.ticker.C:
			go c.Proxy.HeartBeat()
		case <-c.CloseChan:
			c.ticker.Stop()
			if c.Client != nil {
				c.Client.Close()
			}
			return
		}
	}
}

func (p *ServiceProxy) Start() (err error) {
	if err = codec.SetDefaultCodecScheme(codec.CodecSchemeProtobuf); err != nil {
		return err
	}
	if err = p.ServerCommon.Start(); err != nil {
		return err
	}
	p.Agent.Proxy = p
	if err = p.Agent.Start(config.EtcdUrl); err != nil {
		return err
	}
	// Proxy 自己也注册到 etcd，方便其他服务发现代理入口。
	p.AddrServer(&p.info)
	go func() {
		if err = gnet.Run(p, config.ProxyAddr, gnet.WithMulticore(true)); err != nil {
			lib.FatalOnError(err, "Proxy Serve error")
		}
	}()
	go p.Run()
	return
}

func (p *ServiceProxy) OnTraffic(c gnet.Conn) (action gnet.Action) {
	frameCodec := codec.DefaultFrameCodec()
	for {
		frame, err := frameCodec.Decode(c)
		if err == codec.ErrIncompletePacket {
			return
		}
		if err != nil {
			lib.LogErrorAndReturn(err, "ServiceProxy decode frame error")
			return gnet.Close
		}
		p.handleFrame(frame.Body, c)
	}
}

func (p *ServiceProxy) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	p.handleFrame(frame, c)
	return
}

func (p *ServiceProxy) handleFrame(frame []byte, c gnet.Conn) {
	go func() {
		if err := p.workPool.Submit(func() {
			message := &pb.ProtoInternal{}
			if err := codec.Unmarshal(frame, message); err != nil {
				lib.LogErrorAndReturn(err, "ServiceProxy unmarshal message error")
				return
			}
			switch message.Cmd {
			case pb.InternalProxySync:
				// 服务连接到 Proxy 后先发送同步消息，Proxy 才能把连接绑定到服务名。
				p.bindConnection(message.Dst, c)
			case pb.InternalGateToProxy:
				// Gate 发来的业务消息按 Dst 路由到目标 Game。
				dst := message.Dst
				p.mu.RLock()
				service, ok := p.Servers[dst]
				p.mu.RUnlock()
				if ok && service.Type == "game" {
					postMsg := pb.ProtoInternal{
						Cmd:       pb.InternalProxyToGame,
						Dst:       dst,
						SessionId: message.SessionId,
						Data:      message.Data,
					}
					p.MsgChan <- postMsg
				} else {
					lib.SugarLogger.Warnf("ServiceProxy target service not found: %s", dst)
				}
			case pb.InternalProxyToGate:
				// 发回 Gate 的消息按 Dst 路由到目标 Gate。
				dst := message.Dst
				p.mu.RLock()
				service, ok := p.Servers[dst]
				p.mu.RUnlock()
				if ok && service.Type == "gate" {
					p.MsgChan <- *message
				} else {
					lib.SugarLogger.Warnf("ServiceProxy target gate not found: %s", dst)
				}
			}
		}); err != nil {
			lib.LogErrorAndReturn(err, "ServiceProxy submit message error")
		}
	}()
}

func (p *ServiceProxy) Stop() {
	p.stopOnce.Do(func() {
		p.ServerCommon.Stop()
		if p.Agent != nil && p.Agent.CloseChan != nil {
			select {
			case p.Agent.CloseChan <- true:
			default:
			}
		}
		if p.workPool != nil {
			p.workPool.Release()
		}
	})
}

func (p *ServiceProxy) Run() {
	go p.Agent.run(&p.info)
	for {
		select {
		case msg := <-p.MsgChan:
			switch msg.Cmd {
			case pb.InternalProxyToGame:
				p.SendToGame(msg.Dst, msg.SessionId, msg.Data)
			case pb.InternalProxyToGate:
				p.SendToGate(msg.Dst, msg.SessionId, msg.Data)
			default:
				p.SendToService(msg.Dst, msg.Cmd, msg.SessionId, msg.Data)
			}
		case <-p.CloseChan:
			p.Stop()
			return
		}
	}
}

func (p *ServiceProxy) LoadConfig(path string) error {
	return p.ServerCommon.LoadConfig(path)
}

func (p *ServiceProxy) AddrServer(s *ServerInfo) {
	// 如果proxy服务的etcd client不存在，直接退出
	if p.Agent == nil {
		err := errors.New("no etcd agent exist")
		lib.FatalOnError(err, "Register new service")
	}
	// 同时写入 etcd 和本地表；本地表供当前 Proxy 立即路由使用。
	p.Agent.RegisteredSvr <- *s
	p.mu.Lock()
	p.Servers[s.Name] = s
	p.mu.Unlock()
}

func (s *ServiceProxy) HeartBeat() {
	// 心跳阶段先刷新当前服务 JSON，不引入 lease 续约复杂逻辑。
	key := service_common.ServiceKey(s.info.Name)
	if s.Agent == nil || s.Agent.Client == nil {
		return
	}
	value, err := buildServerInfoString(&s.info)
	if err != nil {
		lib.LogIfError(err, "Proxy encode heartbeat info error")
		return
	}
	_, err = s.Agent.Client.Put(context.Background(), key, value)
	lib.LogIfError(err, "Error update workerInfo")
}

func (p *ServiceProxy) bindConnection(name string, c gnet.Conn) {
	if name == "" || c == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// 只有握手后的连接才进入 ServiceConnections。
	p.ServiceConnections[name] = c
	if _, ok := p.Servers[name]; !ok {
		// 如果服务尚未通过 etcd 注册，至少保留一个可路由的本地占位信息。
		host, _, _ := net.SplitHostPort(c.RemoteAddr().String())
		p.Servers[name] = &ServerInfo{Name: name, Type: serviceTypeFromName(name), IP: host}
	}
}

type ServerInfo = service_common.ServerInfo

func NewServerInfo(id int32, ip string, name string, serviceType string, port int32) ServerInfo {
	return service_common.NewServerInfo(id, name, serviceType, ip, port)
}

func serviceTypeFromName(name string) string {
	switch {
	case strings.HasPrefix(name, "game"):
		return "game"
	case strings.HasPrefix(name, "gate"):
		return "gate"
	case strings.HasPrefix(name, "proxy"):
		return "proxy"
	default:
		return ""
	}
}
