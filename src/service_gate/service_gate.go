package service_gate

import (
	"net"
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
	proxyLi   net.Listener
	proxyConn net.Conn
	runChan   chan bool
	h         MessageHandler
	msgChan   chan pb.ProtoInternal
	//*gnet.EventServer
	*gnet.BuiltinEventEngine
}

func NewServiceGate(_name string, id int) *ServiceGate {
	pool, err := ants.NewPool(ants.DefaultAntsPoolSize)
	lib.SysLoggerFatal(err, "New Gate pool error")
	return &ServiceGate{
		workPool: pool,
		wg:       sync.WaitGroup{},
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		msgChan: make(chan pb.ProtoInternal, lib.MaxMessageCount),
		//EventServer:        new(gnet.EventServer),
		BuiltinEventEngine: new(gnet.BuiltinEventEngine),
	}
}

func (s *ServiceGate) SendToGame(buf []byte) {

}

func (s *ServiceGate) SendToLogin(buf []byte) {}

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
	s.ServerCommon.Start()
	go func() {
		s.proxyLi, err = net.Listen("tcp", "127.0.0.1:9001")
		lib.LogErrorAndReturn(err, "Service Gate listen ")
		s.proxyConn, err = s.proxyLi.Accept()
		lib.LogIfError(err, "Accept Proxy error")
	}()
	defer func() {
		s.workPool.Release()
		err := s.gsConn.Close()
		if err != nil {
			return
		}
		err = s.proxyConn.Close()
		if err != nil {
			return
		}
	}()
	go func(gg *ServiceGate) {
		addr := "tcp://" + config.GameGateAddr + ":" + config.GameGatePort
		err = gnet.Run(gg, addr, gnet.WithMulticore(true),
			gnet.WithSocketRecvBuffer(lib.MaxReceiveBufCap),
			gnet.WithLogger(lib.SugarLogger))
		lib.FatalOnError(err, "fatal: start gnet error")
		lib.Log(zap.InfoLevel, "gnet listening", err)
	}(s)

	s.Run()
	return
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
				lib.LogErrorAndReturn(err, "")
				if msg.Dst != s.Name {
					switch msg.Cmd {
					case pb.InternalGateToProxy:
						if strings.Contains(msg.Dst, "proxy") {
							s.SendToProxyMessage(msg)
						}
					case pb.InternalProxyToGate:
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
	defer func() {
		s.workPool.Release()
	}()
	s.wg.Wait()
}

func (s *ServiceGate) Run() {
	for {
		select {
		case msg := <-s.msgChan: // TODO 线程安全
			s.handleMessage(msg)
		case <-s.runChan:
			lib.SugarLogger.Info("running")
		case <-s.CloseChan:
			close(s.runChan)
			close(s.CloseChan)
		}
	}
}

func (s *ServiceGate) handleMessage(msg pb.ProtoInternal) {

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
	err := s.ServerCommon.LoadConfig(path)
	if err != nil {
		lib.LogIfError(err, "SererCommon load configure file error")
		return err
	}
	return nil
}
