package service_proxy

import "testing"

func TestNewServiceProxy(t *testing.T) {
	sp := NewServiceProxy("proxy_1", 1)
	if sp == nil {
		t.Fatal("service proxy is nil")
	}
}
