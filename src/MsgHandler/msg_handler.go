package MsgHandler

import "encoding/binary"

const (
	READ_MESSAGE_INIT_LENGTH = 1024
)

type MessageReadWithHead struct {
	Head    *MessageHead
	Offset  uint32
	Data    []byte
	DataLen int32
}

func NewMessageReadWithHead() *MessageReadWithHead {
	return &MessageReadWithHead{
		Head:   NewMessageHead(),
		Offset: 0,
		Data:   make([]byte, READ_MESSAGE_INIT_LENGTH),
	}
}

// TODO
type Session struct {
}

type IMessageReader interface {
}

type MessageFlag byte

type MessageHead struct {
	flag         MessageFlag
	command      uint32
	bodyLength   uint32
	headerLength uint32
}

func NewMessageHead() *MessageHead {
	return &MessageHead{
		flag:         0,
		command:      0,
		bodyLength:   0,
		headerLength: 9,
	}
}

func (head *MessageHead) Encode(buf []byte) {
	buf[0] = byte(head.flag)
	binary.BigEndian.PutUint32(buf[1:5], head.command)
	binary.BigEndian.PutUint32(buf[5:9], head.bodyLength)
}

func (head *MessageHead) Decode(buf []byte) {
	head.flag = MessageFlag(buf[0])
	head.command = binary.BigEndian.Uint32(buf[1:5])
	head.bodyLength = binary.BigEndian.Uint32(buf[5:9])
}

func (mh *MessageReadWithHead) ReadMessage(session *Session) IMessageReader {
	return MessageReadWithHead{}
}
