package codec

import (
	"errors"
	"fmt"

	gnet "github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

var ErrIncompletePacket = errors.New("incomplete packet")

type Frame struct {
	Head ServerMessageHead
	Body []byte
}

type FrameCodec struct {
	Serializer Serializer
}

func NewFrameCodec(serializer Serializer) FrameCodec {
	if serializer == nil {
		serializer = DefaultSerializer()
	}
	return FrameCodec{Serializer: serializer}
}

func DefaultFrameCodec() FrameCodec {
	return NewFrameCodec(nil)
}

func (fc FrameCodec) serializer() Serializer {
	if fc.Serializer == nil {
		return DefaultSerializer()
	}
	return fc.Serializer
}

func (fc FrameCodec) Encode(cmd uint8, onlineIdx int, msg proto.Message) ([]byte, error) {
	body, err := fc.serializer().Marshal(msg)
	if err != nil {
		return nil, err
	}
	return EncodeFrame(cmd, onlineIdx, body)
}

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
