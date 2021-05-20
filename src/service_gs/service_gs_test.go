package service_gs

import (
	"GoGameServer/src/MsgHandler"
	"GoGameServer/src/protocol"
	"GoGameServer/src/service_common"
	"github.com/panjf2000/ants/v2"
	"github.com/panjf2000/gnet"
	"reflect"
	"testing"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameServer{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				protoFactory: tt.fields.protoFactory,
			}
			gs.Start()
		})
	}
}

func TestGameServer_React(t *testing.T) {
	type fields struct {
		ServerCommon *service_common.ServerCommon
		workPool     *ants.Pool
		protoFactory *protocol.Factory
	}
	type args struct {
		frame []byte
		c     gnet.Conn
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantOut    []byte
		wantAction gnet.Action
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameServer{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				protoFactory: tt.fields.protoFactory,
			}
			gotOut, gotAction := gs.React(tt.args.frame, tt.args.c)
			if !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("React() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
			if gotAction != tt.wantAction {
				t.Errorf("React() gotAction = %v, want %v", gotAction, tt.wantAction)
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameServer{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				protoFactory: tt.fields.protoFactory,
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameServer{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				protoFactory: tt.fields.protoFactory,
			}
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameServer{
				ServerCommon: tt.fields.ServerCommon,
				workPool:     tt.fields.workPool,
				protoFactory: tt.fields.protoFactory,
			}
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
		want *GameServer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGameServer(tt.args._name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGameServer() = %v, want %v", got, tt.want)
			}
		})
	}
}
