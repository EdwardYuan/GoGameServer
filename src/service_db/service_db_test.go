package service_db

import "testing"

func TestNewServiceDB(t *testing.T) {
	s := NewServiceDB("db_1", 1)
	if s == nil {
		t.Fatal("service db is nil")
	}
}
