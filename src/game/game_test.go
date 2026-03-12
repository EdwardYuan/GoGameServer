package game

import "testing"

func TestNewAgentManager(t *testing.T) {
	if NewAgentManager() == nil {
		t.Fatal("expected agent manager")
	}
}
