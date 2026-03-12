package service_game

import "testing"

func TestGameServerStruct(t *testing.T) {
	gs := GameServer{}
	if gs.clients == nil && gs.AgentManager != nil {
		t.Logf("empty server initialized")
	}
}
