package service_gs

import (
	"GoGameServer/src/game"
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
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
