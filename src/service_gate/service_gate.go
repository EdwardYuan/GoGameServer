package service_gate

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
	"go.uber.org/zap"
	"net"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"google.golang.org/protobuf/proto"
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
	go func(gg *ServiceGate) {
		err = gnet.Serve(gg, lib.GNetAddr, gnet.WithMulticore(true),
			gnet.WithCodec(gnet.NewFixedLengthFrameCodec(5)), // gnet.WithCodec(&lib.MsgCodec{}),
			gnet.WithLogger(lib.SugarLogger))
		lib.FatalOnError(err, "fatal: start gnet error")
	}(s)
	li, err := net.Listen("tcp", "127.0.0.1:8890")
	lib.LogIfError(err, "listen to connect error")
	conn, err := li.Accept()
	lib.LogIfError(err, "accept connection error")
	//TODO retry
	if conn != nil {
		s.gsConn = &conn
	}
	s.Run()
	return
}

func (s *ServiceGate) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// lib.Log(zap.DebugLevel, string(frame), nil)
	if s.workPool != nil {
		s.wg.Add(1)
		err := s.workPool.Submit(func() {
			// session := lib.NewSession(c)
			headReader := lib.NewMessageHeadReader()
			headReader.Head.Decode(frame)
			if headReader.Head.Check() != nil {
				return
			}
			headReader.ReadMessage(frame[headReader.Head.HeaderLength:])
			switch headReader.Head.Command {
			case lib.NetMsgToGame:
				s.SendToGame(headReader.Data)
			case lib.NetMsgToLogin:
				s.SendToLogin(headReader.Data)
			case lib.NetMsgToDB:
				s.SendToDB(headReader.Data)
			default:
				return
			}
			var message proto.Message
			// proto.Unmarshal(frame, message)
			if !s.h.Check(message) {
				return
			}
			lib.SugarLogger.Info(message)
			s.wg.Done()
		})
		if err != nil {
			lib.Log(zap.ErrorLevel, "submit message pool error", err)
		}
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