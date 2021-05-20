package service_lg

import (
	"GoGameServer/src/service_common"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := &LoginGate{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				err:          tt.fields.err,
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := &LoginGate{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				err:          tt.fields.err,
			}
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
		want *LoginGate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLoginGate(tt.args._name, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLoginGate() = %v, want %v", got, tt.want)
			}
		})
	}
}
