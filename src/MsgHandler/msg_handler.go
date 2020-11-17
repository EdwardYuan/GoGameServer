package MsgHandler

import (
	"encoding/binary"
	"gogameserver/lib"
	"net"
)

const (
	READ_MESSAGE_INIT_LENGTH = 1024
)

// TODO
type Session struct {
	con net.Conn
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
		flag:         0,
		command:      0,
		bodyLength:   0,
		headerLength: 9,
	}
}

func NewMessageHeadReader() *MessageHeadReader {
	return &MessageHeadReader{
		Head:   NewMessageHead(),
		Offset: 0,
		Data:   make([]byte, READ_MESSAGE_INIT_LENGTH),
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

func (mh *MessageHeadReader) ReadMessage(session *Session) IMessageReader {
	if mh.Offset < mh.Head.headerLength {
		//ead Head
		readNum, err := session.con.Read(mh.Data[mh.Offset:mh.Head.headerLength])
		if err != nil {
			lib.SugarLogger.Errorf("MessageHeadReader ReadMessage err: %+v", err)
			return nil
		}
		mh.Offset += uint32(readNum)
		if mh.Offset < mh.Head.headerLength {
			lib.SugarLogger.Errorf("MessageHeadReader ReadMessage err: Head length read error.")
			return nil
		}
		mh.Head.Decode(mh.Data[:mh.Head.headerLength])
		if mh.Head.bodyLength == 0 {
			lib.SugarLogger.Errorf("MessageHeadReader ReadMessage Err: message body length is zero.")
			return nil
		}
		if mh.Head.bodyLength > mh.MaxDataLen {
			lib.SugarLogger.Errorf("MessageHeadReader ReadMessage Err: too big data.")
			return nil
		}

	}
	return MessageHeadReader{}
}
