package service_common

import "testing"

func TestServerCommonInit(t *testing.T) {
	s := &ServerCommon{Name: "test", Id: 1, CloseChan: make(chan int, 1)}
	if s.Name != "test" {
		t.Fatal("unexpected name")
	}
}
