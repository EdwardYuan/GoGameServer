package MsgHandler

import (
	"GoGameServer/src/lib"
	"encoding/binary"
	"net"
)

const (
	READ_MESSAGE_INIT_LENGTH = 1024
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

func (head *MessageHead) Check() (err error){
	// TODO check flag, command, head length and body length
	err = head.crc()
	return
}

func (head *MessageHead) crc() (err error){
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
		flag:         0,
		command:      0,
		bodyLength:   0,
		headerLength: 0,
	}
}

func (mh *MessageHeadReader) ReadMessage(session *Session) IMessageReader {
	if mh.Offset < mh.Head.headerLength {
		//Read Head
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
		// Read body
		if mh.Offset < mh.Head.headerLength+mh.Head.bodyLength {
			readNum, err := session.con.Read(mh.Data[mh.Offset : mh.Head.headerLength+mh.Head.bodyLength])
			if err != nil {
				lib.SugarLogger.Errorf("MessageHeadReader ReadMessage Err: %+v", err)
				return nil
			}
			mh.Offset += uint32(readNum)
			if mh.Offset < mh.Head.headerLength+mh.Head.bodyLength {
				lib.SugarLogger.Errorf("MessageHeadReader ReadMessage Err: read body not finished.")
				return nil
			} else if mh.Offset == mh.Head.headerLength+mh.Head.bodyLength {
				mh.Offset = 0
				bodyData := make([]byte, mh.Head.bodyLength)
				copy(bodyData, mh.Data[mh.Head.headerLength:mh.Head.headerLength+mh.Head.bodyLength])
				message := newMessage(session, mh.Head.flag, mh.Head.command, bodyData)
				return message
			} else {
				lib.SugarLogger.Errorf("Read message error, should not reach here.")
				return nil
			}
		}
		lib.SugarLogger.Errorf("Read Message Error, read Session %d, read offset %d", session.id, mh.Offset)
		return nil
	}
	return nil
}
