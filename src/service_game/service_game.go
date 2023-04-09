package service_game

import (
	"errors"
	"net"
	"sync"
	"time"

	"GoGameServer/src/codec"
	"GoGameServer/src/game"
	"GoGameServer/src/global"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type GameServer struct {
	runChannel chan bool
	*service_common.ServerCommon
	wg           sync.WaitGroup
	gateConn     net.Conn
	dbConn       net.Conn
	proxyConn    net.Conn
	recvChan     chan protocol.Message
	clients      map[uint64]*Client // gate发过来的SessionId到角色的映射
	AgentManager *game.AgentManager

	// 这是一行注释
	// 下面这部分分离到网络处理中
	readBuffer []byte
	readOffset int
}

func NewGameServer(_name string, id int) *GameServer {
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: &service_common.ServerCommon{
			Name: _name,
			Id:   id,
		},
		wg:           sync.WaitGroup{},
		runChannel:   make(chan bool),
		clients:      make(map[uint64]*Client, lib.MaxOnlineClientCount),
		AgentManager: game.NewAgentManager(),
	}
}

func (gs *GameServer) RegisterService() {
	etcd := viper.Sub("etcd")
	endpoint := etcd.GetString("endpoints")
	global.RegisterService([]string{endpoint}, gs.Name, "game", "")

	// cli, err := clientv3.New(clientv3.Config{
	//	Endpoints:   []string{endpoint}, // TODO 配置多个etcd节点
	//	DialTimeout: 5 * time.Second,
	// })
	// lib.FatalOnError(err, "New Proxy Service error")
	// defer cli.Close()
	// _, err = cli.Put(context.Background(), gs.Name, strconv.Itoa(gs.Id))
	// lib.FatalOnError(err, "Failed to register Service to etcd.")
}

func (gs *GameServer) Start() (err error) {
	gs.ServerCommon.Start()
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	// gs.RegisterService()
	// 连接Gate
	err = gs.connectToGate()
	lib.FatalOnError(err, "Connect to Gate")
	// 连接DBServer
	err = gs.connectToDBServer()
	lib.FatalOnError(err, "Connect to DBServer")
	// 连接Proxy
	err = gs.connectToProxy()
	go gs.netLoop()
	go gs.Run()
	return
}

func (gs *GameServer) connectToProxy() (err error) {
	proxyAddr := viper.GetString("proxy.addr") + viper.GetString("proxy.port")
	gs.proxyConn, err = net.DialTimeout("tcp", proxyAddr, 15*time.Second)
	lib.LogIfError(err, "connect to proxy")
	if gs.proxyConn != nil {
		lib.Log(zap.InfoLevel, "Connected to Proxy successfully", err)
	}
	return
}

func (gs *GameServer) connectToGate() (err error) {
	gateCfg := viper.Sub("gamegate")
	addr := gateCfg.GetString("addr")
	port := gateCfg.GetString("port")
	gs.gateConn, err = net.DialTimeout("tcp", addr+":"+port, 15*time.Second)
	lib.LogIfError(err, "connect to gate error")
	if gs.gateConn != nil {
		lib.Log(zap.InfoLevel, "Connected to GameGate successfully.", err)
	}
	return err
}

func (gs *GameServer) connectToDBServer() (err error) {
	dbConfig := viper.Sub("dbserver")
	addr := dbConfig.GetString("addr")
	port := dbConfig.GetString("port")
	gs.dbConn, err = net.DialTimeout("tcp", addr+":"+port, 15*time.Second)
	lib.LogIfError(err, "connect to db error")
	if gs.dbConn != nil {
		lib.Log(zap.InfoLevel, "Connected to DBServer successfully", err)
	}
	return err
}

func (gs *GameServer) netLoop() {
	for {

		size, err := gs.proxyConn.Read(gs.readBuffer[gs.readOffset:codec.MessageHeadLength])
		lib.LogIfError(err, "GameServer read buffer error")
		if size != codec.MessageHeadLength {
			lib.LogIfError(errors.New("invalid read size"), " GameServer read buffer error")
		}
		head := new(codec.ServerMessageHead)
		head.Decode(gs.readBuffer[:codec.MessageHeadLength])
		if ok, err := head.Check(); err != nil || !ok {
			lib.LogIfError(err, "")
		}
	}

}

func (gs *GameServer) Stop() {
	defer func() {
		close(gs.runChannel)
		err := gs.dbConn.Close()
		if err != nil {
			lib.LogIfError(err, "Close DB Connection Error")
			return
		}
		err = gs.gateConn.Close()
		if err != nil {
			lib.LogIfError(err, "Close Gate Connection Error")
			return
		}
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
		case msg, ok := <-gs.recvChan:
			if ok {
				gs.OnMessageReceived(msg)
			}
		}
	}
}

func (gs *GameServer) OnMessageReceived(msg protocol.Message) {
	protoMessage := &pb.ProtoInternal{}
	switch msg.Command {
	case pb.CMD_INTERNAL_PLAYER_LOGIN:
		// TODO创建session，从消息获取playerid
		client := gs.NewClient(nil, 0)
		err := proto.Unmarshal(msg.Data, protoMessage)
		lib.LogIfError(err, "Unmarshal Message error")
		select {
		case client.Rev <- protoMessage.Data:
			err = client.Start()
			lib.LogIfError(err, "start client error")
		default:
			lib.SugarLogger.Errorf("Player login unmarshal message error %v", err)
		}
	case pb.CMD_INTERNAL_PLAYER_LOGOUT:
		client := gs.clients[msg.SessionId]
		if client != nil {
			// Todo
			err := client.Stop()
			lib.LogIfError(err, "client stop error")
			go func() {
				if client.closed {
					delete(gs.clients, msg.SessionId)
				}
			}()
		}
	case pb.CMD_INTERNAL_PLAYER_TO_GAME_MESSAGE:
		client := gs.clients[msg.SessionId]
		if client != nil {
			err := proto.Unmarshal(msg.Data, protoMessage)
			lib.LogIfError(err, "Unmarshal Message error")
			select {
			case client.Rev <- protoMessage.Data:
				lib.Logger.Info("message received.\n")
			default:
				lib.SugarLogger.Errorf("Player to GameServer message error %v", err)
			}
		}
	}
}
