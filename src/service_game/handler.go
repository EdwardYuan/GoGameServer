package service_game

import (
	"fmt"

	"GoGameServer/network"
	"GoGameServer/src/codec"
	"GoGameServer/src/game"
	"GoGameServer/src/lib"
	"GoGameServer/src/protocol"

	"google.golang.org/protobuf/proto"
)

type handler interface {
	Unmarshal(buf []byte) codec.ServerMessageHead
	Marshal(m proto.Message) []byte
}

type ClientInfo struct{}

type ClientCloseReason int

const (
	ClientCloseNormal ClientCloseReason = iota
	ClientCloseKill
)

// Client  of GameServer
type Client struct {
	info            ClientInfo
	PlayerSessionId uint64
	account         string
	agent           *game.Agent
	Rev             chan []byte
	closeChan       chan ClientCloseReason
	closed          bool
}

func (gs *GameServer) NewClient(session *network.Session, playerId int64) *Client {
	client := new(Client)
	agent := gs.AgentManager.NewAgent(session, playerId)
	client.agent = agent
	gs.clients[client.PlayerSessionId] = client
	return client
}

func (c *Client) Start() (err error) {
	c.run()
	return
}

func (c *Client) Stop() (err error) {
	c.closeChan <- ClientCloseNormal
	return
}

func (c *Client) run() {
	defer close(c.Rev)
	go c.agent.Run()
	for {
		select {
		case data := <-c.Rev:
			var message proto.Message
			err := proto.Unmarshal(data, message)
			lib.LogIfError(err, "unmarshal message error")
			// Todo 反射对应消息处理函数
			se := lib.SeqEvent{}
			if seq, evt, name, _, _, err := protocol.ParseProtobufEvent(data); err != nil {
				lib.SugarLogger.Errorf("handle message error %v", err)
				continue
			} else if evt == nil {
				lib.SugarLogger.Error("handle message event is nil\n")
				continue
			} else {
				fmt.Printf("%s\n", name)
				se.Seq = seq
				se.Event = evt
			}
			eventName := lib.Name(se.Event)
			switch eventName {
			case "C2S_Auth":
			case "C2S_CreateRole":
			case "C2S_DeleteRole":
			case "C2S_EnterGame":
			default:
				if c.agent != nil {
					select {
					case c.agent.Recv <- se:
					case <-c.agent.CloseChan:
						// Todo 退出游戏
					}
				}
			}
		case reason := <-c.closeChan:
			switch reason {
			case ClientCloseNormal:
				c.agent.CloseChan <- 1
				// 连接关闭agent可以继续存在
				c.closed = true
				return
			}
		}
	}
}
