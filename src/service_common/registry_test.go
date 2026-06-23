package service_common

import "testing"

func TestServiceKey(t *testing.T) {
	if got, want := ServiceKey("game_1"), "services/game_1"; got != want {
		t.Fatalf("ServiceKey=%q want %q", got, want)
	}
}

func TestServerInfoMarshalRoundTripIncludesType(t *testing.T) {
	want := &ServerInfo{Id: 1, Name: "gate_1", Type: "gate", IP: "127.0.0.1", Port: 8890}

	encoded, err := MarshalServerInfo(want)
	if err != nil {
		t.Fatalf("MarshalServerInfo failed: %v", err)
	}
	got, err := UnmarshalServerInfo(encoded)
	if err != nil {
		t.Fatalf("UnmarshalServerInfo failed: %v", err)
	}
	if *got != *want {
		t.Fatalf("round trip mismatch: got %+v want %+v", got, want)
	}
}
