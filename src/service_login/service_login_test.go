package service_login

import "testing"

func TestNewLoginGate(t *testing.T) {
	lg := NewLoginGate("login_1", 1)
	if lg == nil {
		t.Fatal("login gate is nil")
	}
}
