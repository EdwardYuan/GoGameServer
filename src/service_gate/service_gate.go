package service_gate

import (
	"net"
	"strconv"
	"strings"
	"sync"

	"GoGameServer/src/codec"
	"GoGameServer/src/config"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"go.uber.org/zap"

	"github.com/panjf2000/ants/v2"
	gnet "github.com/panjf2000/gnet/v2"
)

type ServiceGate struct {
	err      error
	workPool *ants.Pool
	wg       sync.WaitGroup
	*service_common.ServerCommon
	gsConn    net.Conn
	proxyConn net.Conn
	h         MessageHandler
	msgChan   chan pb.ProtoInternal
	stopOnce  sync.Once
	//*gnet.EventServer
	*gnet.BuiltinEventEngine
}

func NewServiceGate(_name string, id int) *ServiceGate {
	pool, err := ants.NewPool(ants.DefaultAntsPoolSize)
	lib.SysLoggerFatal(err, "New Gate pool error")
	return &ServiceGate{
		workPool:     pool,
		wg:           sync.WaitGroup{},
		ServerCommon: service_common.NewServerCommon(_name, id),
		msgChan:      make(chan pb.ProtoInternal, lib.MaxMessageCount),
		//EventServer:        new(gnet.EventServer),
		BuiltinEventEngine: new(gnet.BuiltinEventEngine),
	}
}

func (s *ServiceGate) SendToGame(buf []byte) {
	lib.SugarLogger.Debugf("ServiceGate SendToGame is not wired yet, dropped %d bytes", len(buf))
}

func (s *ServiceGate) SendToLogin(buf []byte) {
	lib.SugarLogger.Debugf("ServiceGate SendToLogin is not wired yet, dropped %d bytes", len(buf))
}

// SendToDB 不一定有用，暂时不需要gate直接和dbserver交互
func (s *ServiceGate) SendToDB(buf []byte) {}

func (s *ServiceGate) Error() string {
	if s.err != nil {
		lib.SugarLogger.Error(s.err)
		return "GameGate Error"
	}
	return ""
}

func (s *ServiceGate) Start() (err error) {
	if err = codec.SetDefaultCodecScheme(codec.CodecSchemeProtobuf); err != nil {
		return err
	}
	lib.SugarLogger.Info("Service Gate Start: ", s.Name)
	if err = s.ServerCommon.Start(); err != nil {
		return err
	}
	go func(gg *ServiceGate) {
		addr := "tcp://" + config.GameGateAddr + ":" + config.GameGatePort
		err = gnet.Run(gg, addr, gnet.WithMulticore(true),
			gnet.WithSocketRecvBuffer(lib.MaxReceiveBufCap),
			gnet.WithLogger(lib.SugarLogger))
		lib.FatalOnError(err, "fatal: start gnet error")
		lib.Log(zap.InfoLevel, "gnet listening", err)
	}(s)
	if err = s.connectToProxy(); err != nil {
		return err
	}
	// Gate 对外接收客户端连接，启动成功后注册到 etcd，供 Proxy 回包路由。
	if err = s.registerService(); err != nil {
		return err
	}
	// 主动连接 Proxy 后发送握手，Proxy 才能把该连接绑定到 gate 名称。
	if err = s.syncProxyConnection(); err != nil {
		return err
	}

	go s.Run()
	return
}

func (s *ServiceGate) connectToProxy() error {
	conn, err := net.Dial("tcp", config.ProxyAddr)
	if err != nil {
		return err
	}
	s.proxyConn = conn
	return nil
}

// registerService 注册 Gate 的对外服务地址；db/login 当前不走该注册路径。
func (s *ServiceGate) registerService() error {
	port, err := strconv.Atoi(config.GameGatePort)
	if err != nil {
		return err
	}
	info := service_common.NewServerInfo(int32(s.Id), s.Name, "gate", lib.GetLocalIP(lib.IPv4), int32(port))
	return service_common.RegisterService(config.EtcdUrl, info)
}

// syncProxyConnection 发送 InternalProxySync 握手，让 Proxy 记录 gateName -> conn。
func (s *ServiceGate) syncProxyConnection() error {
	if s.proxyConn == nil {
		return net.ErrClosed
	}
	msg := &pb.ProtoInternal{
		Cmd: pb.InternalProxySync,
		Dst: s.Name,
	}
	packet, err := codec.EncodeMessage(msg)
	if err != nil {
		return err
	}
	_, err = s.proxyConn.Write(packet)
	return err
}

func (s *ServiceGate) OnTraffic(c gnet.Conn) (action gnet.Action) {
	frameCodec := codec.DefaultFrameCodec()
	for {
		frame, err := frameCodec.Decode(c)
		if err == codec.ErrIncompletePacket {
			return
		}
		if err != nil {
			lib.LogErrorAndReturn(err, "ServiceGate decode frame error")
			return gnet.Close
		}
		s.handleFrame(frame.Body)
	}
}

func (s *ServiceGate) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	s.handleFrame(frame)
	return
}

func (s *ServiceGate) handleFrame(frame []byte) {
	var err error
	if s.workPool == nil {
		s.workPool, err = ants.NewPool(global.DefaultPoolSize)
		lib.LogErrorAndReturn(err, "service gate new pool error")
	}
	if s.workPool != nil && err == nil {
		s.wg.Add(1)
		go func() {
			err := s.workPool.Submit(func() {
				msg := &pb.ProtoInternal{}
				err = codec.Unmarshal(frame, msg)
				if lib.LogErrorAndReturn(err, "") {
					return
				}
				switch msg.Cmd {
				case pb.InternalGateToProxy:
					// 目标是 Proxy 的消息直接写入已握手的 proxyConn。
					if strings.Contains(msg.Dst, "proxy") {
						s.SendToProxyMessage(msg)
					}
				case pb.InternalProxyToGate:
					// 目标是当前 Gate 的消息进入本地处理队列。
					if msg.Dst == "" || msg.Dst == s.Name {
						s.msgChan <- *msg
					}
				}
			})
			if err != nil {
				lib.Log(zap.ErrorLevel, "submit message pool error", err)
			}
			s.wg.Done()
		}()
	}
}

func (s *ServiceGate) Stop() {
	s.stopOnce.Do(func() {
		s.ServerCommon.Stop()
		if s.gsConn != nil {
			if err := s.gsConn.Close(); err != nil {
				lib.LogIfError(err, "ServiceGate close game connection error")
			}
		}
		if s.proxyConn != nil {
			if err := s.proxyConn.Close(); err != nil {
				lib.LogIfError(err, "ServiceGate close proxy connection error")
			}
		}
		if s.workPool != nil {
			s.workPool.Release()
		}
		s.wg.Wait()
	})
}

func (s *ServiceGate) Run() {
	for {
		select {
		case msg := <-s.msgChan: // TODO 线程安全
			s.handleMessage(msg)
		case <-s.CloseChan:
			s.Stop()
			return
		}
	}
}

func (s *ServiceGate) handleMessage(msg pb.ProtoInternal) {
	switch msg.Cmd {
	case pb.InternalProxyToGate:
		s.SendToGame(msg.Data)
	default:
		lib.SugarLogger.Debugf("ServiceGate received unsupported message cmd=%d dst=%s", msg.Cmd, msg.Dst)
	}
}

func (s *ServiceGate) SendToProxy(data []byte) {
	if s.proxyConn != nil {
		_, err := s.proxyConn.Write(data)
		if err != nil {
			return
		}
	}
}

func (s *ServiceGate) SendToProxyMessage(msg *pb.ProtoInternal) {
	if s.proxyConn == nil || msg == nil {
		return
	}
	packet, err := codec.EncodeMessage(msg)
	if err != nil {
		lib.LogErrorAndReturn(err, "ServiceGate encode proxy message error")
		return
	}
	if _, err := s.proxyConn.Write(packet); err != nil {
		lib.LogErrorAndReturn(err, "ServiceGate send proxy message error")
	}
}

func (s *ServiceGate) LoadConfig(path string) error {
	return s.ServerCommon.LoadConfig(path)
}
