package service_game

import (
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/service_common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"
	"time"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	wg       sync.WaitGroup
	gateConn net.Conn
	dbConn   net.Conn
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

func (gs *GameServer) RegisterService() {
	etcd := viper.Sub("etcd")
	endpoint := etcd.GetString("endpoints")
	global.RegisterService([]string{endpoint}, gs.Name, "game", "")

	//cli, err := clientv3.New(clientv3.Config{
	//	Endpoints:   []string{endpoint}, // TODO 配置多个etcd节点
	//	DialTimeout: 5 * time.Second,
	//})
	//lib.FatalOnError(err, "New Proxy Service error")
	//defer cli.Close()
	//_, err = cli.Put(context.Background(), gs.Name, strconv.Itoa(gs.Id))
	//lib.FatalOnError(err, "Failed to register Service to etcd.")
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	//gs.RegisterService()
	// 连接Gate
	err = gs.connectToGate()
	lib.FatalOnError(err, "Connect to Gate")
	// 连接DBServer
	err = gs.connectToDBServer()
	lib.FatalOnError(err, "Connect to DBServer")
	gs.Run()
	return
}

func (gs *GameServer) connectToGate() (err error) {
	gatecfg := viper.Sub("gamegate")
	addr := gatecfg.GetString("addr")
	port := gatecfg.GetString("port")
	gs.gateConn, err = net.DialTimeout("tcp", addr+":"+port, 15*time.Second)
	lib.LogIfError(err, "connect to gate error")
	if gs.gateConn != nil {
		lib.Log(zap.InfoLevel, "Connected to GameGate successfully.", nil)
	}
	return err
}

func (gs *GameServer) connectToDBServer() (err error) {
	dbcfg := viper.Sub("dbserver")
	addr := dbcfg.GetString("addr")
	port := dbcfg.GetString("port")
	gs.dbConn, err = net.DialTimeout("tcp", addr+":"+port, 15*time.Second)
	lib.LogIfError(err, "connect to db error")
	if gs.dbConn != nil {
		lib.Log(zap.InfoLevel, "Connected to DBServer successfully", nil)
	}
	return err
}

func (gs *GameServer) netLoop() {

}

func (gs *GameServer) Stop() {
	defer func() {
		close(gs.runChannel)
		gs.dbConn.Close()
		gs.gateConn.Close()
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
		lib.LogIfError(err, "Unmarshal Message error")
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
			lib.LogIfError(err, "Unmarshal Message error")
			select {
			case client.Recv <- protoMessage.Data:
				lib.Logger.Info("message received.\n")
			default:
				lib.SugarLogger.Errorf("Player to GameServer message error %v", err)
			}
		}
	}
}
