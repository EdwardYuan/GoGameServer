package service_common

import (
	"github.com/panjf2000/gnet"
	"gogameserver/MsgHandler"
)

type Service interface {
	Start(_name string, args []interface{}) (err error)
	Stop()
	Run()
}

type ServerCommon struct {
	*gnet.EventServer
}

func (s *ServerCommon) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	return
}

func (s *ServerCommon) Encode(data []byte) (msg MsgHandler.Message, err error) {
	return
}

func (s *ServerCommon) Decode(msg MsgHandler.Message) (err error) {
	return
}
