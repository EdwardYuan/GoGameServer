package service_db

import (
	"net"
	"sync"

	"GoGameServer/src/config"
	"GoGameServer/src/lib"
	"GoGameServer/src/service_common"
)

type ServiceDB struct {
	gsConn net.Conn
	li     net.Listener
	*service_common.ServerCommon
	stopOnce sync.Once
}

func NewServiceDB(_name string, idx int) *ServiceDB {
	return &ServiceDB{
		ServerCommon: service_common.NewServerCommon(_name, idx),
	}
}

func (s *ServiceDB) Start() error {
	if err := s.ServerCommon.Start(); err != nil {
		return err
	}
	addr := config.DBServiceAddr + ":" + config.DBServicePort
	li, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.li = li
	go func() {
		conn, err := s.li.Accept()
		if err != nil {
			lib.LogIfError(err, "accept")
			return
		}
		s.gsConn = conn
	}()
	go s.Run()
	return err
}

func (s *ServiceDB) Stop() {
	s.stopOnce.Do(func() {
		s.ServerCommon.Stop()
		if s.gsConn != nil {
			if err := s.gsConn.Close(); err != nil {
				lib.LogIfError(err, "stop service db connection error")
			}
		}
		if s.li != nil {
			if err := s.li.Close(); err != nil {
				lib.LogIfError(err, "stop service db listener error")
			}
		}
	})
}

func (s *ServiceDB) Run() {
	for {
		select {
		case <-s.CloseChan:
			s.Stop()
			return
		}
	}
}
