package global

import "testing"

func TestServerMapAddress(t *testing.T) {
	makeSvcStringMap()
	sm := NewServerMapAddress()
	sm.MapAddrToServerName("127.0.0.1", "game", "game_1")
	if typ := sm.GetSvrTypeByAddr("127.0.0.1"); typ != ServerGame {
		t.Fatalf("expected %v got %v", ServerGame, typ)
	}
}
