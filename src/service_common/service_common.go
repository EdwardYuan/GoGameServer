package service_common

import (
	"github.com/panjf2000/gnet"
	"gogameserver/MsgHandler"
)

type Service struct {
	*gnet.EventServer
}

func (s *Service) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {

	return
}

func (s *Service) Encode(data []byte) (msg MsgHandler.Message, err error) {
	return
}

func (s *Service) Decode(msg MsgHandler.Message) (err error) {
	return
}
