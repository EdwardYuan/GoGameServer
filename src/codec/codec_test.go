package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"GoGameServer/src/pb"
	"google.golang.org/protobuf/proto"
)

type fakeSerializer struct {
	marshalCalled   bool
	unmarshalCalled bool
}

func (f *fakeSerializer) Name() string {
	return "fake"
}

func (f *fakeSerializer) Marshal(proto.Message) ([]byte, error) {
	f.marshalCalled = true
	return []byte("fake-data"), nil
}

func (f *fakeSerializer) Unmarshal([]byte, proto.Message) error {
	f.unmarshalCalled = true
	return nil
}

func TestDefaultSerializerIsProtobuf(t *testing.T) {
	if err := SetDefaultCodecScheme(CodecSchemeProtobuf); err != nil {
		t.Fatal(err)
	}

	msg := &pb.ProtoInternal{
		Cmd:       pb.InternalGateToProxy,
		Dst:       "game-1",
		SessionId: 42,
		Data:      []byte("payload"),
	}
	data, err := Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	got := &pb.ProtoInternal{}
	if err := Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	if got.Cmd != msg.Cmd || got.Dst != msg.Dst || got.SessionId != msg.SessionId || !bytes.Equal(got.Data, msg.Data) {
		t.Fatalf("unexpected decoded message: %+v", got)
	}
}

func TestSetDefaultSerializer(t *testing.T) {
	defer SetDefaultCodecScheme(CodecSchemeProtobuf)

	fake := &fakeSerializer{}
	if err := SetDefaultSerializer(fake); err != nil {
		t.Fatal(err)
	}
	data, err := Marshal(&pb.ProtoInternal{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fake-data" || !fake.marshalCalled {
		t.Fatalf("custom serializer was not used")
	}
	if err := Unmarshal(data, &pb.ProtoInternal{}); err != nil {
		t.Fatal(err)
	}
	if !fake.unmarshalCalled {
		t.Fatalf("custom serializer unmarshal was not used")
	}
}

func TestSetDefaultSerializerNil(t *testing.T) {
	if err := SetDefaultCodecScheme(CodecSchemeProtobuf); err != nil {
		t.Fatal(err)
	}
	before := DefaultSerializer()
	err := SetDefaultSerializer(nil)
	if !errors.Is(err, ErrNilSerializer) {
		t.Fatalf("got %v, want %v", err, ErrNilSerializer)
	}
	if DefaultSerializer() != before {
		t.Fatalf("nil serializer replaced default serializer")
	}
}

func TestSetDefaultCodecScheme(t *testing.T) {
	defer SetDefaultCodecScheme(CodecSchemeProtobuf)

	if err := SetDefaultCodecScheme(CodecSchemeMsg); err != nil {
		t.Fatal(err)
	}
	if DefaultSerializer().Name() != string(CodecSchemeMsg) {
		t.Fatalf("got %q, want %q", DefaultSerializer().Name(), CodecSchemeMsg)
	}

	if err := SetDefaultCodecScheme(CodecSchemeProtobuf); err != nil {
		t.Fatal(err)
	}
	if DefaultSerializer().Name() != string(CodecSchemeProtobuf) {
		t.Fatalf("got %q, want %q", DefaultSerializer().Name(), CodecSchemeProtobuf)
	}
}

func TestSetDefaultCodecSchemeUnknown(t *testing.T) {
	defer SetDefaultCodecScheme(CodecSchemeProtobuf)

	if err := SetDefaultCodecScheme(CodecSchemeMsg); err != nil {
		t.Fatal(err)
	}
	before := DefaultSerializer()
	if err := SetDefaultCodecScheme("unknown"); err == nil {
		t.Fatalf("expected unsupported codec scheme error")
	}
	if DefaultSerializer() != before {
		t.Fatalf("unsupported codec scheme replaced default serializer")
	}
}

func TestMsgSerializerRoundTrip(t *testing.T) {
	serializer := MsgSerializer{}
	msg := &pb.ProtoInternal{
		Cmd:       pb.InternalGateToProxy,
		Dst:       "proxy-1",
		SessionId: 77,
		Data:      []byte("payload"),
	}

	data, err := serializer.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	got := &pb.ProtoInternal{}
	if err := serializer.Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	if got.Cmd != msg.Cmd || got.Dst != msg.Dst || got.SessionId != msg.SessionId || !bytes.Equal(got.Data, msg.Data) {
		t.Fatalf("unexpected decoded message: %+v", got)
	}
}

func TestMsgSerializerEmptyFields(t *testing.T) {
	serializer := MsgSerializer{}
	msg := &pb.ProtoInternal{}

	data, err := serializer.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	got := &pb.ProtoInternal{Data: []byte("existing")}
	if err := serializer.Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	if got.Cmd != 0 || got.Dst != "" || got.SessionId != 0 || len(got.Data) != 0 {
		t.Fatalf("unexpected decoded empty message: %+v", got)
	}
}

func TestMsgSerializerRejectsUnsupportedMessage(t *testing.T) {
	serializer := MsgSerializer{}

	if _, err := serializer.Marshal(&pb.Person{}); err == nil {
		t.Fatalf("expected marshal error for unsupported message")
	}
	if err := serializer.Unmarshal(nil, &pb.Person{}); err == nil {
		t.Fatalf("expected unmarshal error for unsupported message")
	}
}

func TestMsgSerializerRejectsOversizedDst(t *testing.T) {
	serializer := MsgSerializer{}
	msg := &pb.ProtoInternal{
		Dst: strings.Repeat("a", msgMaxDstLength+1),
	}

	if _, err := serializer.Marshal(msg); err == nil {
		t.Fatalf("expected oversized dst error")
	}
}

func TestMsgSerializerRejectsMalformedData(t *testing.T) {
	serializer := MsgSerializer{}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "short packet",
			data: make([]byte, msgMinLength-1),
		},
		{
			name: "dst length exceeds buffer",
			data: func() []byte {
				data := make([]byte, msgMinLength)
				binary.LittleEndian.PutUint16(data[msgCmdLength+msgSessionIDLength:msgCmdLength+msgSessionIDLength+msgDstLengthLength], 1)
				return data
			}(),
		},
		{
			name: "body length exceeds buffer",
			data: func() []byte {
				data := make([]byte, msgMinLength)
				binary.LittleEndian.PutUint32(data[msgCmdLength+msgSessionIDLength+msgDstLengthLength:msgMinLength], 1)
				return data
			}(),
		},
		{
			name: "trailing bytes",
			data: func() []byte {
				data, err := serializer.Marshal(&pb.ProtoInternal{Data: []byte("payload")})
				if err != nil {
					t.Fatal(err)
				}
				return append(data, 0)
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := serializer.Unmarshal(tt.data, &pb.ProtoInternal{}); err == nil {
				t.Fatalf("expected malformed data error")
			}
		})
	}
}

func TestDefaultFrameCodecUsesSelectedScheme(t *testing.T) {
	defer SetDefaultCodecScheme(CodecSchemeProtobuf)
	if err := SetDefaultCodecScheme(CodecSchemeMsg); err != nil {
		t.Fatal(err)
	}

	msg := &pb.ProtoInternal{
		Cmd:       pb.InternalGateToProxy,
		Dst:       "proxy-1",
		SessionId: 77,
		Data:      []byte("payload"),
	}
	packet, err := DefaultFrameCodec().Encode(uint8(msg.Cmd), 0, msg)
	if err != nil {
		t.Fatal(err)
	}
	frame, _, err := DecodeFrame(packet)
	if err != nil {
		t.Fatal(err)
	}
	got := &pb.ProtoInternal{}
	if err := Unmarshal(frame.Body, got); err != nil {
		t.Fatal(err)
	}
	if got.Cmd != msg.Cmd || got.Dst != msg.Dst || got.SessionId != msg.SessionId || !bytes.Equal(got.Data, msg.Data) {
		t.Fatalf("unexpected decoded message: %+v", got)
	}
}

func TestServerMessageHeadRoundTrip(t *testing.T) {
	head := &ServerMessageHead{
		Flag:       1,
		PieceFlag:  2,
		Cmd:        3,
		DataLength: 1024,
		OnLineIdx:  99,
	}
	buf := make([]byte, MessageHeadLength)
	if err := head.EncodeTo(buf); err != nil {
		t.Fatal(err)
	}

	var got ServerMessageHead
	got.Decode(buf)
	if got != *head {
		t.Fatalf("got %+v, want %+v", got, *head)
	}
}

func TestDecodeFrame(t *testing.T) {
	body := []byte("hello")
	packet, err := EncodeFrame(7, 11, body)
	if err != nil {
		t.Fatal(err)
	}
	frame, consumed, err := DecodeFrame(packet)
	if err != nil {
		t.Fatal(err)
	}
	if consumed != len(packet) {
		t.Fatalf("consumed %d, want %d", consumed, len(packet))
	}
	if frame.Head.Cmd != 7 || frame.Head.OnLineIdx != 11 || !bytes.Equal(frame.Body, body) {
		t.Fatalf("unexpected frame: %+v", frame)
	}
}

func TestDecodeFrameIncomplete(t *testing.T) {
	if _, _, err := DecodeFrame(make([]byte, MessageHeadLength-1)); !errors.Is(err, ErrIncompletePacket) {
		t.Fatalf("got %v, want %v", err, ErrIncompletePacket)
	}

	packet, err := EncodeFrame(1, 0, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := DecodeFrame(packet[:len(packet)-1]); !errors.Is(err, ErrIncompletePacket) {
		t.Fatalf("got %v, want %v", err, ErrIncompletePacket)
	}
}

func TestDecodeFrameTooLong(t *testing.T) {
	buf := make([]byte, MessageHeadLength)
	binary.LittleEndian.PutUint64(buf[3:11], uint64(ServerMaxReceiveLength+1))

	if _, _, err := DecodeFrame(buf); err == nil {
		t.Fatalf("expected oversized frame error")
	}
}
