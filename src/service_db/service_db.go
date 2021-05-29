package service_db

import "GoGameServer/src/service_common"

type ServiceDB struct {
	*service_common.ServerCommon
}

func NewServiceDB() *ServiceDB {
	return &ServiceDB{
		&service_common.ServerCommon{
			Name:        "",
			Id:          0,
			CloseChan:   nil,
			Rabbit:      nil,
			SvrTick:     nil,
			EventServer: nil,
		},
	}
}

func (s *ServiceDB) Start() error {
	return nil
}

func (s *ServiceDB) Stop() {

}

func (s *ServiceDB) Run() {

}
