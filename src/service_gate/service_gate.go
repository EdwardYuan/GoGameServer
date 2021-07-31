package service_gate

import "GoGameServer/src/service_common"

type ServiceGate struct {
	*service_common.ServerCommon
}

func NewServiceGate(_name string, id int) *ServiceGate {
	return &ServiceGate{ServerCommon: &service_common.ServerCommon{
		Name: _name,
		Id:   id,
	},
	}
}

func (s *ServiceGate) Start() (err error) {
	return
}

func (s *ServiceGate) Stop() {

}

func (s *ServiceGate) Run() {

}

func (s *ServiceGate) LoadConfig(path string) error {
	return nil
}
