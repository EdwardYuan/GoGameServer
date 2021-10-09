package lib

import (
	"encoding/binary"

	"github.com/panjf2000/gnet"
)

const (
	NetMsgToGame = iota + 1
	NetMsgToLogin
	NetMsgToDB
)

const (
	ReadMessageInitLength = 1024
	MESSAGE_COMMAND_INDEX = 5
	MESSAGE_FLAG_INDEX    = 9

	MaxMessageBodySize = 4 * 1024
)

// TODO
type Session struct {
	id  uint64
	con gnet.Conn
}

func NewSession(c gnet.Conn) *Session {
	return &Session{
		id:  genSessionId(),
		con: c,
	}
}

func genSessionId() uint64 {
	return 0
}

type IMessageHandler interface {
	Encode(msg Message) (data []byte, err error)
	Decode(data []byte) (msg Message, err error)
}

type IMessageReader interface {
}

type MessageFlag byte

type MessageHead struct {
	Flag         MessageFlag
	Command      uint32
	BodyLength   uint32
	HeaderLength uint32
}

type MessageHeadReader struct {
	Head       *MessageHead
	Offset     uint32
	Data       []byte
	MaxDataLen uint32
}

type MessageHeadWriter struct {
	Head    *MessageHead
	Offset  uint32
	Buf     []byte
	DataLen int32
}

func NewMessageHead() *MessageHead {
	return &MessageHead{
		Flag:         0,
		Command:      0,
		BodyLength:   0,
		HeaderLength: 9,
	}
}

func NewMessageHeadReader() *MessageHeadReader {
	return &MessageHeadReader{
		Head:   NewMessageHead(),
		Offset: 0,
		Data:   make([]byte, ReadMessageInitLength),
	}
}

func (head *MessageHead) Encode(buf []byte) {
	buf[0] = byte(head.Flag)
	binary.BigEndian.PutUint32(buf[1:5], head.Command)
	binary.BigEndian.PutUint32(buf[5:9], head.BodyLength)
}

func (head *MessageHead) Decode(buf []byte) {
	head.Flag = MessageFlag(buf[0])
	head.Command = binary.BigEndian.Uint32(buf[1:MESSAGE_COMMAND_INDEX])
	head.BodyLength = binary.BigEndian.Uint32(buf[MESSAGE_COMMAND_INDEX:MESSAGE_FLAG_INDEX])
}

func (head *MessageHead) Check() (err error) {
	// TODO check flag, command, head length and body length
	err = head.crc()
	return
}

func (head *MessageHead) crc() (err error) {
	return
}

//接收和发送的消息结构
type Message struct {
	Flag      MessageFlag //标志位
	Command   uint32      //消息类型
	SessionId uint64      //会话
	Data      []byte
}

func newMessage(session *Session, flag MessageFlag, command uint32, data []byte) *Message {
	message := &Message{}
	message.Flag = flag
	message.Command = command
	message.SessionId = session.id
	message.Data = data
	return message
}

func (mh *MessageHeadReader) ReadMessage(buf []byte) IMessageReader {
	mh.Data = buf
	return nil
}
