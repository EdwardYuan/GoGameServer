package tests

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"GoGameServer/src/service_gs"
	"reflect"
	"testing"

	"github.com/panjf2000/ants/v2"
)

func TestGameServer_AddMessageNode(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		protoFactory *protocol.Factory
	}
	type args struct {
		msg *MsgHandler.Message
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := service_gs.NewGameServer(tt.name, idx)
			gs.Start()
		})
	}
}

func TestGameServer_Run(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		protoFactory *protocol.Factory
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := service_gs.NewGameServer(tt.name, idx)
			gs.Start()
		})
	}
}

func TestGameServer_Start(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		protoFactory *protocol.Factory
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := service_gs.NewGameServer(tt.name, idx)
			if err := gs.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGameServer_Stop(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		protoFactory *protocol.Factory
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := service_gs.NewGameServer(tt.name, idx)
			gs.Start()

		})
	}
}

func TestNewGameServer(t *testing.T) {
	type args struct {
		_name string
	}
	tests := []struct {
		name string
		args args
		want *service_gs.GameServer
	}{
		// TODO: Add test cases.
	}
	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service_gs.NewGameServer(tt.args._name, idx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGameServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
