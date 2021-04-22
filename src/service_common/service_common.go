package service_common

import (
	"GoGameServer/src/MsgHandler"
	"github.com/panjf2000/gnet"
)

type Service interface {
	Start() (err error)
	Stop()
	Run()
}

type ServerCommon struct {
	Name      string
	Id        int
	CloseChan chan int
	*gnet.EventServer
}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	return
}

func (s *ServerCommon) Encode(msg MsgHandler.Message) (err error) {
	//TODO protobuf marshal
	return
}

func (s *ServerCommon) Decode(data []byte) (msg MsgHandler.Message, err error) {
	//TODO protobuf unmarshal
	return
}

func (s *ServerCommon) HandleMessage(msg MsgHandler.Message) {

}
