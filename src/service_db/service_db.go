package service_db

import "GoGameServer/src/service_common"

type ServiceDB struct {
	*service_common.ServerCommon
}

func NewServiceDB(_name string, idx int) *ServiceDB {
	return &ServiceDB{
		&service_common.ServerCommon{
			Name:        _name,
			Id:          idx,
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
