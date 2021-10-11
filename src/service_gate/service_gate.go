package service_gate

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
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
	go func(gg *ServiceGate) {
		gnet.Serve(gg, lib.GNetAddr, gnet.WithMulticore(true),
			gnet.WithCodec(gnet.NewFixedLengthFrameCodec(5)), // gnet.WithCodec(&lib.MsgCodec{}),
			gnet.WithLogger(lib.SugarLogger))
	}(s)
	s.Run()
	return
}

func (s *ServiceGate) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// lib.Log(zap.DebugLevel, string(frame), nil)
	if s.workPool != nil {
		s.wg.Add(1)
		s.workPool.Submit(func() {
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
	return nil
}
