package service_game

import (
	"GoGameServer/src/game"
	"GoGameServer/src/lib"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"fmt"
	"google.golang.org/protobuf/proto"
)

type handler interface {
	Unmarshal(buf []byte) lib.Message
	Marshal(m lib.Message) []byte
}

type ClientInfo struct{}

type ClientCloseReason int

type Client struct {
	info            ClientInfo
	PlayerSessionId uint64
	account         string
	service         service_common.Service
	agent           *game.Agent
	Recv            chan []byte
	closeChan       chan ClientCloseReason
}

func (gs *GameServer) NewClient() *Client {
	client := new(Client)
	gs.clients[client.PlayerSessionId] = client
	return client
}

func (c *Client) run() {
	for {
		select {
		case data := <-c.Recv:
			var message proto.Message
			err := proto.Unmarshal(data, message)
			lib.LogIfError(err, "unmarshal message error")
			//Todo 反射对应消息处理函数
			se := lib.SeqEvent{}
			if seq, evt, name, _, _, err := protocol.ParseProtobufEvent(data); err != nil {
				lib.SugarLogger.Errorf("handle message error %v", err)
				continue
			} else if evt == nil {
				lib.SugarLogger.Error("handle message event is nil\n")
				continue
			} else {
				fmt.Printf(name)
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
						//Todo 退出游戏
					}
				}
			}
		}
	}
}
