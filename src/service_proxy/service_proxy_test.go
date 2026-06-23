package service_proxy

import "testing"

func TestServerInfoJSONRoundTrip(t *testing.T) {
	want := &ServerInfo{Id: 1, Name: "game_1", Type: "game", IP: "127.0.0.1", Port: 8890}

	encoded, err := buildServerInfoString(want)
	if err != nil {
		t.Fatalf("buildServerInfoString failed: %v", err)
	}
	got := makeServerInfo(encoded)
	if got == nil {
		t.Fatal("makeServerInfo returned nil")
	}
	if *got != *want {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, want)
	}
}

func TestAddrServerUsesServiceNameKey(t *testing.T) {
	p := NewServiceProxy("proxy_0", 0)
	info := &ServerInfo{Id: 1, Name: "game_1", Type: "game", IP: "127.0.0.1", Port: 8890}

	p.AddrServer(info)

	if _, ok := p.Servers["game_1"]; !ok {
		t.Fatalf("Servers missing service name key: %+v", p.Servers)
	}
}

func TestServiceTypeFromName(t *testing.T) {
	cases := map[string]string{
		"game_1":  "game",
		"gate_1":  "gate",
		"proxy_0": "proxy",
		"other_1": "",
	}
	for name, want := range cases {
		if got := serviceTypeFromName(name); got != want {
			t.Fatalf("serviceTypeFromName(%q)=%q want %q", name, got, want)
		}
	}
}
