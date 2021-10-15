package service_gs

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"sync"

	"google.golang.org/protobuf/proto"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	wg       sync.WaitGroup
	recvChan chan proto.Message
	clients  map[uint64]*Client
}

func NewGameServer(_name string, id int) *GameServer {
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		wg:         sync.WaitGroup{},
		runChannel: make(chan bool),
	}
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	gs.Run()
	return
}

func (gs *GameServer) Stop() {
	defer func() {
		close(gs.runChannel)
	}()
	gs.wg.Wait()
	gs.runChannel <- false
	lib.SugarLogger.Info("Service ", gs.Name, " Stopped.")
}

func (gs *GameServer) Run() {
	for {
		select {
		case <-gs.CloseChan:
			gs.Stop()
		case <-gs.runChannel:
			lib.SugarLogger.Info("running...")
		}
	}
}

func (gs *GameServer) OnMessageReceived(msg lib.Message) {
	protoMessage := &pb.ProtoInternal{}
	switch msg.Command {
	case pb.CMD_INTERNAL_PLAYER_LOGIN:
		client := gs.NewClient()
		err := proto.Unmarshal(msg.Data, protoMessage)
		select {
		case client.Recv <- protoMessage.Data:
		default:
			lib.SugarLogger.Errorf("Player login unmarshal message error %v", err)
		}
	case pb.CMD_INTERNAL_PLAYER_LOGOUT:
		client := gs.clients[msg.SessionId]
		if client != nil {
			// Todo
			delete(gs.clients, msg.SessionId)
		}
	case pb.CMD_INTERNAL_PLAYER_TO_GAME_MESSAGE:
		client := gs.clients[msg.SessionId]
		if client != nil {
			err := proto.Unmarshal(msg.Data, protoMessage)
			select {
			case client.Recv <- protoMessage.Data:
				lib.Logger.Info("message received.\n")
			default:
				lib.SugarLogger.Errorf("Player to GameServer message error %v", err)
			}
		}
	}
}
