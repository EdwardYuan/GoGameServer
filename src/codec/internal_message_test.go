package codec

import (
	"testing"

	"GoGameServer/src/pb"
)

func TestEncodeMessageDecodeFrameRoundTrip(t *testing.T) {
	if err := SetDefaultCodecScheme(CodecSchemeProtobuf); err != nil {
		t.Fatalf("SetDefaultCodecScheme failed: %v", err)
	}
	want := &pb.ProtoInternal{
		Cmd:       pb.InternalGateToProxy,
		Dst:       "game_1",
		SessionId: 42,
		Data:      []byte("payload"),
	}

	packet, err := EncodeMessage(want)
	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}
	frame, consumed, err := DecodeFrame(packet)
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}
	if consumed != len(packet) {
		t.Fatalf("consumed=%d want %d", consumed, len(packet))
	}
	got := &pb.ProtoInternal{}
	if err := Unmarshal(frame.Body, got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got.Cmd != want.Cmd || got.Dst != want.Dst || got.SessionId != want.SessionId || string(got.Data) != string(want.Data) {
		t.Fatalf("message mismatch: got %+v want %+v", got, want)
	}
}
