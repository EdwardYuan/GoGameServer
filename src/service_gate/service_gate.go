package service_gate

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
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
	*gnet.EventServer
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
		EventServer: new(gnet.EventServer),
	}
}

func (s *ServiceGate) SendToGame(buf []byte) {

}

func (s *ServiceGate) SendToLogin(buf []byte) {}

// 不一定有用，暂时不需要gate直接和dbserver交互
func (s *ServiceGate) SendToDB(buf []byte) {}

func (s *ServiceGate) Error() string {
	if s.err != nil {
		lib.SugarLogger.Error(s.err)
		return "GameGate Error"
	}
	return ""
}

func (s *ServiceGate) Start() (err error) {
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
		s.gsConn.Close()
		s.proxyConn.Close()
	}()
	go func(gg *ServiceGate) {
		err = gnet.Serve(gg, lib.GNetAddr, gnet.WithMulticore(true),
			gnet.WithCodec(codec.MsgCodec{}),
			//gnet.WithCodec(gnet.NewFixedLengthFrameCodec(5)),
			//gnet.WithCodec(codec.CodecProtobuf{}),
			gnet.WithLogger(lib.SugarLogger))
		lib.FatalOnError(err, "fatal: start gnet error")
		lib.Log(zap.InfoLevel, "gnet listening", err)
	}(s)

	s.Run()
	return
}

func (s *ServiceGate) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	var err error
	if s.workPool == nil {
		s.workPool, err = ants.NewPool(global.DefaultPoolSize)
		lib.LogIfError(err, "service gate new pool error")
	}
	if s.workPool != nil && err == nil {
		s.wg.Add(1)
		go func() {
			err := s.workPool.Submit(func() {
				// session := lib.NewSession(c)
				//headReader := lib.NewMessageHeadReader()
				//headReader.Head.Decode(frame)
				//if headReader.Head.Check() != nil {
				//	return
				//}
				//headReader.ReadMessage(frame[headReader.Head.HeaderLength:])
				//switch headReader.Head.Command {
				//case lib.NetMsgToGame:
				//	s.SendToGame(headReader.Data)
				//case lib.NetMsgToLogin:
				//	s.SendToLogin(headReader.Data)
				//case lib.NetMsgToDB:
				//	s.SendToDB(headReader.Data)
				//default:
				//	return
				//}

				/*
					//var message proto.Message
					msg := &pb.Person1{}
					err := proto.Unmarshal(frame, msg)
					lib.LogIfError(err, "unmarshal message error")
					if !s.h.Check(msg) {
						return
					}
					lib.SugarLogger.Info(msg.Id)
					lib.SugarLogger.Info(msg.Name)
					lib.SugarLogger.Info(msg.Email)
				*/
				msg := &pb.ProtoInternal{}
				err = proto.Unmarshal(frame, msg)
				lib.LogErrorAndReturn(err, "")
				switch msg.Cmd {
				case pb.InternalGateToProxy:
					s.SendToProxy(frame)
				}
				//s.SendToProxy(msg)
			})
			if err != nil {
				lib.Log(zap.ErrorLevel, "submit message pool error", err)
			}
			s.wg.Done()
		}()
		s.wg.Wait()
	}
	return
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
		case <-s.runChan:
			lib.SugarLogger.Info("running")
		case <-s.CloseChan:
			close(s.runChan)
			close(s.CloseChan)
		}
	}
}

func (s *ServiceGate) SendToProxy(data []byte) {
	if s.proxyConn != nil {
		s.proxyConn.Write(data)
	}
}

func (s *ServiceGate) LoadConfig(path string) error {
	s.ServerCommon.LoadConfig(path)
	return nil
}
