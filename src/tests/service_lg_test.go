package tests

import (
	"GoGameServer/src/service_common"
	"GoGameServer/src/service_lg"
	"github.com/panjf2000/ants/v2"
	"reflect"
	"testing"
)

func TestLoginGate_Close(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		err          error
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := service_lg.NewLoginGate(tt.name, idx)
			lg.StartRabbit()
		})
	}
}

func TestLoginGate_run(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		err          error
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := service_lg.NewLoginGate(tt.name, idx)
			lg.Start()
		})
	}
}

func TestNewLoginGate(t *testing.T) {
	type args struct {
		_name string
		id    int
	}
	tests := []struct {
		name string
		args args
		want *service_lg.LoginGate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service_lg.NewLoginGate(tt.args._name, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLoginGate() = %v, want %v", got, tt.want)
			}
		})
	}
}
