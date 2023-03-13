package service_db

import (
	"net"
	"time"

	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
)

type ServiceDB struct {
	gsConn net.Conn
	*service_common.ServerCommon
}

func NewServiceDB(_name string, idx int) *ServiceDB {
	return &ServiceDB{
		ServerCommon: &service_common.ServerCommon{
			Name:      _name,
			Id:        idx,
			CloseChan: make(chan int, 1),
			SvrTick:   time.NewTicker(10 * time.Millisecond),
		},
	}
}

func (s *ServiceDB) Start() error {
	li, err := net.Listen("tcp", "127.0.0.1:8891")
	lib.FatalOnError(err, "listen to gameserver")
	if li != nil {
		s.gsConn, err = li.Accept()
		lib.LogIfError(err, "accept")
	}
	s.Run()
	return err
}

func (s *ServiceDB) Stop() {
	defer func(gsConn net.Conn) {
		err := gsConn.Close()
		if err != nil {
			lib.LogIfError(err, "stop service db error")
		}
	}(s.gsConn)
}

func (s *ServiceDB) Run() {
}
