package service_gate

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"fmt"
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
	gsConn  *net.Conn
	runChan chan bool
	h       MessageHandler
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
	defer s.workPool.Release()
	go func(gg *ServiceGate) {
		err = gnet.Serve(gg, lib.GNetAddr, gnet.WithMulticore(true),
			gnet.WithCodec(codec.MsgCodec{}),
			//gnet.WithCodec(gnet.NewFixedLengthFrameCodec(5)),
			//gnet.WithCodec(codec.CodecProtobuf{}),
			gnet.WithLogger(lib.SugarLogger))
		lib.FatalOnError(err, "fatal: start gnet error")
		lib.Log(zap.InfoLevel, "gnet listening", err)
	}(s)

	//li, err := net.Listen("tcp", "127.0.0.1:8890")
	//lib.LogIfError(err, "listen to connect error")
	//conn, err := li.Accept()
	//lib.LogIfError(err, "accept connection error")
	//TODO retry
	//if conn != nil {
	//	s.gsConn = &conn
	//}
	s.Run()
	return
}

func (s *ServiceGate) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	lib.SugarLogger.Infof("len of frame is %d", len(frame))
	lib.SugarLogger.Infof(fmt.Sprintf("data is %s", frame))
	out = frame

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
				//var message proto.Message
				var msg *pb.Person1
				err := proto.Unmarshal(frame, msg)
				lib.LogIfError(err, "unmarshal message error")
				if !s.h.Check(msg) {
					return
				}
				lib.SugarLogger.Info(msg)
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

func (s *ServiceGate) LoadConfig(path string) error {
	s.ServerCommon.LoadConfig(path)
	return nil
}
