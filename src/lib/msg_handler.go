package lib

import (
	"encoding/binary"
	"net"
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
	con net.Conn
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

func newMessageHead(buf []byte, len int) *MessageHead {
	return &MessageHead{
		Flag:         0,
		Command:      0,
		BodyLength:   0,
		HeaderLength: 9,
	}
}

func (mh *MessageHeadReader) ReadMessage(session *Session) IMessageReader {
	if mh.Offset < mh.Head.HeaderLength {
		//Read Head
		readNum, err := session.con.Read(mh.Data[mh.Offset:mh.Head.HeaderLength])
		if err != nil {
			SugarLogger.Errorf("MessageHeadReader ReadMessage err: %+v", err)
			return nil
		}
		mh.Offset += uint32(readNum)
		if mh.Offset < mh.Head.HeaderLength {
			SugarLogger.Errorf("MessageHeadReader ReadMessage err: Head length read error.")
			return nil
		}
		mh.Head.Decode(mh.Data[:mh.Head.HeaderLength])
		if mh.Head.BodyLength == 0 {
			SugarLogger.Errorf("MessageHeadReader ReadMessage Err: message body length is zero.")
			return nil
		}
		if mh.Head.BodyLength > mh.MaxDataLen {
			SugarLogger.Errorf("MessageHeadReader ReadMessage Err: too big data.")
			return nil
		}
		// Read body
		if mh.Offset < mh.Head.HeaderLength+mh.Head.BodyLength {
			readNum, err := session.con.Read(mh.Data[mh.Offset : mh.Head.HeaderLength+mh.Head.BodyLength])
			if err != nil {
				SugarLogger.Errorf("MessageHeadReader ReadMessage Err: %+v", err)
				return nil
			}
			mh.Offset += uint32(readNum)
			if mh.Offset < mh.Head.HeaderLength+mh.Head.BodyLength {
				SugarLogger.Errorf("MessageHeadReader ReadMessage Err: read body not finished.")
				return nil
			} else if mh.Offset == mh.Head.HeaderLength+mh.Head.BodyLength {
				mh.Offset = 0
				bodyData := make([]byte, mh.Head.BodyLength)
				copy(bodyData, mh.Data[mh.Head.HeaderLength:mh.Head.HeaderLength+mh.Head.BodyLength])
				message := newMessage(session, mh.Head.Flag, mh.Head.Command, bodyData)
				return message
			} else {
				SugarLogger.Errorf("Read message error, should not reach here.")
				return nil
			}
		}
		SugarLogger.Errorf("Read Message Error, read Session %d, read offset %d", session.id, mh.Offset)
		return nil
	}
	return nil
}
