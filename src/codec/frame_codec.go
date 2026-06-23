// frame_codec.go 提供服务间 TCP 流的帧级编解码能力。
//
// 帧格式由固定长度 ServerMessageHead 加可变长度 Body 组成；
// Body 的序列化方式由 Serializer 决定，默认使用 protobuf。
package codec

import (
	"errors"
	"fmt"

	gnet "github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

// ErrIncompletePacket 表示当前缓冲区还不足以解析出一个完整消息帧。
var ErrIncompletePacket = errors.New("incomplete packet")

// Frame 表示已经从 TCP 流中拆出的一个完整消息帧。
type Frame struct {
	Head ServerMessageHead
	Body []byte
}

// FrameCodec 负责消息帧的编码、拆包和包体序列化策略选择。
type FrameCodec struct {
	Serializer Serializer
}

// NewFrameCodec 创建一个帧编解码器。
//
// serializer 为 nil 时使用当前全局默认 Serializer。
func NewFrameCodec(serializer Serializer) FrameCodec {
	if serializer == nil {
		serializer = DefaultSerializer()
	}
	return FrameCodec{Serializer: serializer}
}

// DefaultFrameCodec 使用当前默认 Serializer 创建帧编解码器。
func DefaultFrameCodec() FrameCodec {
	return NewFrameCodec(nil)
}

// serializer 返回当前 FrameCodec 实际使用的 Serializer。
func (fc FrameCodec) serializer() Serializer {
	if fc.Serializer == nil {
		return DefaultSerializer()
	}
	return fc.Serializer
}

// Encode 将 protobuf 消息序列化并封装成带包头的完整消息帧。
func (fc FrameCodec) Encode(cmd uint8, onlineIdx int, msg proto.Message) ([]byte, error) {
	body, err := fc.serializer().Marshal(msg)
	if err != nil {
		return nil, err
	}
	return EncodeFrame(cmd, onlineIdx, body)
}

// EncodeFrame 将已序列化的消息体封装成服务间消息帧。
func EncodeFrame(cmd uint8, onlineIdx int, body []byte) ([]byte, error) {
	head := ServerMessageHead{
		Cmd:        cmd,
		DataLength: len(body),
		OnLineIdx:  onlineIdx,
	}
	if _, err := head.Check(); err != nil {
		return nil, err
	}
	out := make([]byte, MessageHeadLength+len(body))
	if err := head.EncodeTo(out[:MessageHeadLength]); err != nil {
		return nil, err
	}
	copy(out[MessageHeadLength:], body)
	return out, nil
}

// DecodeFrame 从字节切片中解析一个完整消息帧。
//
// 返回值 consumed 表示本次消费的字节数；如果数据不足，返回 ErrIncompletePacket。
func DecodeFrame(data []byte) (Frame, int, error) {
	if len(data) < MessageHeadLength {
		return Frame{}, 0, ErrIncompletePacket
	}
	var head ServerMessageHead
	head.Decode(data[:MessageHeadLength])
	if _, err := head.Check(); err != nil {
		return Frame{}, 0, err
	}
	frameLen := MessageHeadLength + head.DataLength
	if len(data) < frameLen {
		return Frame{}, 0, ErrIncompletePacket
	}
	body := make([]byte, head.DataLength)
	copy(body, data[MessageHeadLength:frameLen])
	return Frame{Head: head, Body: body}, frameLen, nil
}

// Decode 从 gnet 连接缓冲区中读取并消费一个完整消息帧。
//
// 数据不足时不会消费缓冲区，并返回 ErrIncompletePacket。
func (fc FrameCodec) Decode(c gnet.Conn) (Frame, error) {
	headBuf, err := c.Peek(MessageHeadLength)
	if err != nil || len(headBuf) < MessageHeadLength {
		return Frame{}, ErrIncompletePacket
	}
	var head ServerMessageHead
	head.Decode(headBuf)
	if _, err := head.Check(); err != nil {
		return Frame{}, err
	}
	frameLen := MessageHeadLength + head.DataLength
	if c.InboundBuffered() < frameLen {
		return Frame{}, ErrIncompletePacket
	}
	frameBuf, err := c.Peek(frameLen)
	if err != nil || len(frameBuf) < frameLen {
		return Frame{}, ErrIncompletePacket
	}
	body := make([]byte, head.DataLength)
	copy(body, frameBuf[MessageHeadLength:frameLen])
	discarded, err := c.Discard(frameLen)
	if err != nil {
		return Frame{}, err
	}
	if discarded != frameLen {
		return Frame{}, fmt.Errorf("discarded %d bytes, want %d", discarded, frameLen)
	}
	return Frame{Head: head, Body: body}, nil
}
