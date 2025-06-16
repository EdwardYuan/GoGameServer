package service_gate

import "testing"

func TestNewServiceGate(t *testing.T) {
	sg := NewServiceGate("gate_1", 1)
	if sg == nil {
		t.Fatal("service gate is nil")
	}
}
