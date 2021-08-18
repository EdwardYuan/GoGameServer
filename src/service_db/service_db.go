package service_db

import (
	"GoGameServer/src/service_common"
	"time"
)

type ServiceDB struct {
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
	s.Run()
	return nil
}

func (s *ServiceDB) Stop() {

}

func (s *ServiceDB) Run() {

}
