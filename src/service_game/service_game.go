package service_game

import (
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"GoGameServer/src/codec"
	"GoGameServer/src/config"
	"GoGameServer/src/game"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type GameServer struct {
	*service_common.ServerCommon
	wg           sync.WaitGroup
	gateConn     net.Conn
	dbConn       net.Conn
	proxyConn    net.Conn
	recvChan     chan protocol.Message
	clients      map[uint64]*Client // gate发过来的SessionId到角色的映射
	clientsMu    sync.RWMutex
	AgentManager *game.AgentManager

	// 这是一行注释
	// 下面这部分分离到网络处理中
	readBuffer []byte
	readOffset int
	stopOnce   sync.Once
}

func NewGameServer(_name string, id int) *GameServer {
	lib.SugarLogger.Info("Service ", _name, " created")
	return &GameServer{
		ServerCommon: service_common.NewServerCommon(_name, id),
		wg:           sync.WaitGroup{},
		recvChan:     make(chan protocol.Message, lib.MaxMessageCount),
		clients:      make(map[uint64]*Client, lib.MaxOnlineClientCount),
		AgentManager: game.NewAgentManager(),
		readBuffer:   make([]byte, 0, lib.MaxReceiveBufCap),
	}
}

func (gs *GameServer) Start() (err error) {
	if err = codec.SetDefaultCodecScheme(codec.CodecSchemeProtobuf); err != nil {
		return err
	}
	if err = gs.ServerCommon.Start(); err != nil {
		return err
	}
	lib.SugarLogger.Info("Service ", gs.Name, " Start...")
	// 连接Gate
	err = gs.connectToGate()
	if err != nil {
		return err
	}
	// 连接DBServer
	err = gs.connectToDBServer()
	if err != nil {
		return err
	}
	// 连接Proxy
	err = gs.connectToProxy()
	if err != nil {
		return err
	}
	// 基础依赖连接成功后再注册自身，避免 etcd 中出现不可用的 game 实例。
	if err = gs.registerService(); err != nil {
		return err
	}
	// 通知 Proxy 当前 TCP 连接对应的服务名，供后续路由使用。
	if err = gs.syncProxyConnection(); err != nil {
		return err
	}
	go gs.netLoop()
	go gs.Run()
	return
}

func (gs *GameServer) registerService() error {
	port, err := strconv.Atoi(config.GameServerPort)
	if err != nil {
		return err
	}
	info := service_common.NewServerInfo(int32(gs.Id), gs.Name, "game", lib.GetLocalIP(lib.IPv4), int32(port))
	return service_common.RegisterService(config.EtcdUrl, info)
}

// syncProxyConnection 发送 InternalProxySync 握手，让 Proxy 将连接绑定到 gs.Name。
func (gs *GameServer) syncProxyConnection() error {
	if gs.proxyConn == nil {
		return io.ErrClosedPipe
	}
	msg := &pb.ProtoInternal{
		Cmd: pb.InternalProxySync,
		Dst: gs.Name,
	}
	packet, err := codec.EncodeMessage(msg)
	if err != nil {
		return err
	}
	_, err = gs.proxyConn.Write(packet)
	return err
}

func (gs *GameServer) connectToProxy() (err error) {
	proxyAddr := viper.GetString("proxy.addr")
	if port := viper.GetString("proxy.port"); port != "" {
		proxyAddr = proxyAddr + ":" + port
	}
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
	if gs.proxyConn == nil {
		lib.LogIfError(io.ErrClosedPipe, "GameServer proxy connection is nil")
		return
	}
	buf := make([]byte, lib.MaxReceiveBufCap)
	for {
		n, err := gs.proxyConn.Read(buf)
		if err != nil {
			lib.LogIfError(err, "GameServer read proxy error")
			return
		}
		gs.readBuffer = append(gs.readBuffer, buf[:n]...)
		for {
			// TCP 是流式协议，这里从累积缓冲区中拆出完整帧。
			frame, consumed, err := codec.DecodeFrame(gs.readBuffer)
			if err == codec.ErrIncompletePacket {
				break
			}
			if err != nil {
				lib.LogIfError(err, "GameServer decode proxy frame error")
				gs.readBuffer = gs.readBuffer[:0]
				break
			}
			gs.readBuffer = gs.readBuffer[consumed:]
			internal := &pb.ProtoInternal{}
			if err := codec.Unmarshal(frame.Body, internal); err != nil {
				lib.LogIfError(err, "GameServer unmarshal internal message error")
				continue
			}
			msg := protocol.Message{
				SessionId: internal.SessionId,
				Command:   uint32(internal.Cmd),
				Data:      internal.Data,
			}
			// 网络 goroutine 只投递消息，业务处理收敛到 Run 循环。
			select {
			case gs.recvChan <- msg:
			case <-gs.CloseChan:
				return
			}
		}
	}
}

func (gs *GameServer) Stop() {
	gs.stopOnce.Do(func() {
		gs.ServerCommon.Stop()
		if gs.dbConn != nil {
			if err := gs.dbConn.Close(); err != nil {
				lib.LogIfError(err, "Close DB Connection Error")
			}
		}
		if gs.gateConn != nil {
			if err := gs.gateConn.Close(); err != nil {
				lib.LogIfError(err, "Close Gate Connection Error")
			}
		}
		if gs.proxyConn != nil {
			if err := gs.proxyConn.Close(); err != nil {
				lib.LogIfError(err, "Close Proxy Connection Error")
			}
		}
		gs.wg.Wait()
		lib.SugarLogger.Info("Service ", gs.Name, " Stopped.")
	})
}

func (gs *GameServer) Run() {
	for {
		select {
		case <-gs.CloseChan:
			gs.Stop()
			return
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
		gs.clientsMu.Lock()
		gs.clients[msg.SessionId] = client
		gs.clientsMu.Unlock()
		select {
		case client.Rev <- protoMessage.Data:
			err = client.Start()
			lib.LogIfError(err, "start client error")
		default:
			lib.SugarLogger.Errorf("Player login unmarshal message error %v", err)
		}
	case pb.CMD_INTERNAL_PLAYER_LOGOUT:
		gs.clientsMu.RLock()
		client := gs.clients[msg.SessionId]
		gs.clientsMu.RUnlock()
		if client != nil {
			// Todo
			err := client.Stop()
			lib.LogIfError(err, "client stop error")
			go func() {
				if client.closed {
					gs.clientsMu.Lock()
					delete(gs.clients, msg.SessionId)
					gs.clientsMu.Unlock()
				}
			}()
		}
	case pb.CMD_INTERNAL_PLAYER_TO_GAME_MESSAGE:
		gs.clientsMu.RLock()
		client := gs.clients[msg.SessionId]
		gs.clientsMu.RUnlock()
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
