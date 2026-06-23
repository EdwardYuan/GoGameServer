// codec_test.go 验证 codec 包的帧协议、序列化器切换和兼容接口行为。
package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"GoGameServer/src/pb"
	gnet "github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

type fakeSerializer struct {
	marshalCalled   bool
	unmarshalCalled bool
}

// fakeConn 是用于单元测试 FrameCodec/MsgCodec 的最小 gnet.Conn 实现。
type fakeConn struct {
	in []byte
}

var _ gnet.Conn = (*fakeConn)(nil)

// Name 返回测试序列化器的名称。
func (f *fakeSerializer) Name() string {
	return "fake"
}

// Marshal 记录调用并返回固定测试数据。
func (f *fakeSerializer) Marshal(proto.Message) ([]byte, error) {
	f.marshalCalled = true
	return []byte("fake-data"), nil
}

// Unmarshal 记录调用并直接成功返回。
func (f *fakeSerializer) Unmarshal([]byte, proto.Message) error {
	f.unmarshalCalled = true
	return nil
}

// Read 从 fakeConn 的入站缓冲区读取数据。
func (c *fakeConn) Read(p []byte) (int, error) {
	if len(c.in) == 0 {
		return 0, io.EOF
	}
	n := copy(p, c.in)
	c.in = c.in[n:]
	return n, nil
}

// Write 模拟写入成功。
func (c *fakeConn) Write(p []byte) (int, error) { return len(p), nil }

// WriteTo 将 fakeConn 当前缓冲区写入目标 writer。
func (c *fakeConn) WriteTo(w io.Writer) (int64, error) { n, err := w.Write(c.in); return int64(n), err }

// ReadFrom 丢弃从 reader 读到的数据。
func (c *fakeConn) ReadFrom(r io.Reader) (int64, error) { return io.Copy(io.Discard, r) }

// Next 返回并消费后续 n 字节数据。
func (c *fakeConn) Next(n int) ([]byte, error) { p, err := c.Peek(n); c.Discard(len(p)); return p, err }

// Peek 返回后续 n 字节数据但不消费缓冲区。
func (c *fakeConn) Peek(n int) ([]byte, error) {
	if len(c.in) < n {
		return c.in, io.ErrUnexpectedEOF
	}
	return c.in[:n], nil
}

// Discard 从入站缓冲区丢弃 n 字节。
func (c *fakeConn) Discard(n int) (int, error) {
	if len(c.in) < n {
		d := len(c.in)
		c.in = nil
		return d, io.ErrUnexpectedEOF
	}
	c.in = c.in[n:]
	return n, nil
}

// InboundBuffered 返回当前可读缓冲区长度。
func (c *fakeConn) InboundBuffered() int { return len(c.in) }

// Writev 模拟批量写入成功并返回总字节数。
func (c *fakeConn) Writev(bs [][]byte) (int, error) {
	n := 0
	for _, b := range bs {
		n += len(b)
	}
	return n, nil
}

// Flush 模拟刷新写缓冲区成功。
func (c *fakeConn) Flush() error { return nil }

// OutboundBuffered 返回模拟写缓冲区长度。
func (c *fakeConn) OutboundBuffered() int { return 0 }

// AsyncWrite 模拟异步写入成功。
func (c *fakeConn) AsyncWrite([]byte, gnet.AsyncCallback) error { return nil }

// AsyncWritev 模拟批量异步写入成功。
func (c *fakeConn) AsyncWritev([][]byte, gnet.AsyncCallback) error {
	return nil
}

// Fd 返回模拟文件描述符。
func (c *fakeConn) Fd() int { return 0 }

// Dup 模拟复制文件描述符。
func (c *fakeConn) Dup() (int, error) { return 0, nil }

// SetReadBuffer 模拟设置读缓冲区成功。
func (c *fakeConn) SetReadBuffer(int) error { return nil }

// SetWriteBuffer 模拟设置写缓冲区成功。
func (c *fakeConn) SetWriteBuffer(int) error { return nil }

// SetLinger 模拟设置 linger 参数成功。
func (c *fakeConn) SetLinger(int) error { return nil }

// SetKeepAlivePeriod 模拟设置 keepalive 周期成功。
func (c *fakeConn) SetKeepAlivePeriod(time.Duration) error { return nil }

// SetNoDelay 模拟设置 TCP_NODELAY 成功。
func (c *fakeConn) SetNoDelay(bool) error { return nil }

// Context 返回测试连接上下文。
func (c *fakeConn) Context() interface{} { return nil }

// SetContext 设置测试连接上下文。
func (c *fakeConn) SetContext(interface{}) {}

// LocalAddr 返回测试本地地址。
func (c *fakeConn) LocalAddr() net.Addr { return nil }

// RemoteAddr 返回测试远端地址。
func (c *fakeConn) RemoteAddr() net.Addr { return nil }

// SetDeadline 模拟设置连接截止时间成功。
func (c *fakeConn) SetDeadline(time.Time) error { return nil }

// SetReadDeadline 模拟设置读截止时间成功。
func (c *fakeConn) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline 模拟设置写截止时间成功。
func (c *fakeConn) SetWriteDeadline(time.Time) error { return nil }

// Wake 模拟唤醒连接成功。
func (c *fakeConn) Wake(gnet.AsyncCallback) error { return nil }

// CloseWithCallback 模拟带回调关闭连接成功。
func (c *fakeConn) CloseWithCallback(gnet.AsyncCallback) error { return nil }

// Close 模拟关闭连接成功。
func (c *fakeConn) Close() error { return nil }

// TestDefaultSerializerIsProtobuf 验证默认序列化器使用 protobuf。
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

// TestSetDefaultSerializer 验证可以替换全局默认序列化器。
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

// TestSetDefaultSerializerNil 验证 nil 序列化器不会覆盖默认值。
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

// TestSetDefaultCodecScheme 验证内置序列化协议可以按名称切换。
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

// TestSetDefaultCodecSchemeUnknown 验证未知协议名称不会改变当前默认序列化器。
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

// TestMsgSerializerRoundTrip 验证自定义 ProtoInternal 编码可以往返。
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

// TestMsgSerializerEmptyFields 验证自定义编码能正确处理空字段。
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

// TestMsgSerializerRejectsUnsupportedMessage 验证自定义编码拒绝非 ProtoInternal 消息。
func TestMsgSerializerRejectsUnsupportedMessage(t *testing.T) {
	serializer := MsgSerializer{}

	if _, err := serializer.Marshal(&pb.Person{}); err == nil {
		t.Fatalf("expected marshal error for unsupported message")
	}
	if err := serializer.Unmarshal(nil, &pb.Person{}); err == nil {
		t.Fatalf("expected unmarshal error for unsupported message")
	}
}

// TestMsgSerializerRejectsOversizedDst 验证 Dst 超过编码上限时返回错误。
func TestMsgSerializerRejectsOversizedDst(t *testing.T) {
	serializer := MsgSerializer{}
	msg := &pb.ProtoInternal{
		Dst: strings.Repeat("a", msgMaxDstLength+1),
	}

	if _, err := serializer.Marshal(msg); err == nil {
		t.Fatalf("expected oversized dst error")
	}
}

// TestMsgSerializerRejectsMalformedData 验证畸形自定义消息体会被拒绝。
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

// TestDefaultFrameCodecUsesSelectedScheme 验证默认帧编解码器使用当前选择的序列化协议。
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

// TestEncodeMessageRejectsInvalidInput 验证内部消息编码会拒绝 nil 和越界命令号。
func TestEncodeMessageRejectsInvalidInput(t *testing.T) {
	defer SetDefaultCodecScheme(CodecSchemeProtobuf)
	if err := SetDefaultCodecScheme(CodecSchemeProtobuf); err != nil {
		t.Fatal(err)
	}

	if _, err := EncodeMessage(nil); err == nil {
		t.Fatalf("expected nil message error")
	}
	if _, err := EncodeMessage(&pb.ProtoInternal{Cmd: -1}); err == nil {
		t.Fatalf("expected negative cmd error")
	}
	if _, err := EncodeMessage(&pb.ProtoInternal{Cmd: 256}); err == nil {
		t.Fatalf("expected oversized cmd error")
	}
}

// TestDecodeData 验证旧 DecodeData 接口能从完整帧中提取消息体。
func TestDecodeData(t *testing.T) {
	body := []byte("payload")
	packet, err := EncodeFrame(9, 12, body)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := DecodeData(packet)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Cmd != 9 || msg.Dst != "" || msg.SessionId != 0 || !bytes.Equal(msg.Data, body) {
		t.Fatalf("unexpected decoded data: %+v", msg)
	}
}

// TestDecodeDataIncomplete 验证 DecodeData 对短包返回 ErrIncompletePacket。
func TestDecodeDataIncomplete(t *testing.T) {
	if _, err := DecodeData(make([]byte, MessageHeadLength-1)); !errors.Is(err, ErrIncompletePacket) {
		t.Fatalf("got %v, want %v", err, ErrIncompletePacket)
	}

	packet, err := EncodeFrame(1, 0, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeData(packet[:len(packet)-1]); !errors.Is(err, ErrIncompletePacket) {
		t.Fatalf("got %v, want %v", err, ErrIncompletePacket)
	}
}

// TestMsgCodecDecodeUsesFrameCodec 验证 MsgCodec.Decode 复用 FrameCodec 拆包逻辑。
func TestMsgCodecDecodeUsesFrameCodec(t *testing.T) {
	body := []byte("payload")
	packet, err := EncodeFrame(7, 11, body)
	if err != nil {
		t.Fatal(err)
	}

	got, err := (MsgCodec{}).Decode(&fakeConn{in: packet})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("got %q, want %q", got, body)
	}
}

// TestServerMessageHeadRoundTrip 验证服务间消息头可以编码后再解析回来。
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

// TestDecodeFrame 验证完整帧可以被正确解析。
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

// TestDecodeFrameIncomplete 验证 DecodeFrame 对头部或 body 不完整的输入返回短包错误。
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

// TestDecodeFrameTooLong 验证超过最大消息体长度的帧会被拒绝。
func TestDecodeFrameTooLong(t *testing.T) {
	buf := make([]byte, MessageHeadLength)
	binary.LittleEndian.PutUint64(buf[3:11], uint64(ServerMaxReceiveLength+1))

	if _, _, err := DecodeFrame(buf); err == nil {
		t.Fatalf("expected oversized frame error")
	}
}
