package protocol

import (
	"GoGameServer/src/lib"
	"encoding/binary"
	"errors"
	"reflect"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	SessionId uint64
	Command   uint32
	ID        int32
	Body      proto.Message
	Data      []byte
}

func (m *Message) Type() string {
	return string(proto.MessageName(m.Body))
}

var protoTypesNils map[string]proto.Message

func nameToType(name string) (reflect.Type, bool) {
	if t, ok := protoTypesNils[name]; ok {
		return reflect.TypeOf(t), true
	}
	msgType := proto.MessageType(name)
	if msgType == nil {
		return nil, false
	}
	return msgType, true
}

func NewMessageProto(id int32, pm proto.Message) *Message {
	return &Message{
		ID:   id,
		Body: pm,
	}
}

var (
	ErrTooShort             = errors.New("bad message")
	ErrHeaderLengthOverflow = errors.New("header length overflow")
)

var (
	TypeToCodeDict = map[string]int32{}
	CodeToTypeDict = map[int32]string{}
)

func init() {
	// 初始化 TypeToCodeDict 和 CodeToTypeDict
}

func Decode(bs []byte) (msg *Message, headerLen, bodyLen int, err error) {
	totalLen := len(bs)
	if totalLen < 2 {
		err = ErrTooShort
		return
	}
	headerLen = int(binary.LittleEndian.Uint16(bs[:2]))
	if 2+headerLen > totalLen {
		err = ErrHeaderLengthOverflow
		return
	}
	bodySlice := bs[2+headerLen:]
	bodyLen = totalLen - 2 - headerLen

	// 假设这里需要从 header 中提取 msgID
	msgID := int32(0) // 这里需要从 header 中获取实际的 msgID
	name := CodeToTypeDict[msgID]
	t, ok := nameToType(name)
	if !ok {
		err = errors.New("message type not supported for decoding")
		return
	}

	v := reflect.New(t.Elem())
	pm := v.Interface().(proto.Message)
	if err = proto.Unmarshal(bodySlice, pm); err != nil {
		lib.SugarLogger.Warnf("Unmarshal failed name %s", name)
		return
	}
	msg = &Message{ID: msgID, Body: pm}
	return
}

func Encode(m *Message) ([]byte, string, int, int, error) {
	return m.Encode()
}

func (m *Message) Encode() (out []byte, typeStr string, headerLen, bodyLen int, err error) {
	bodySlice, err := proto.Marshal(m.Body)
	if err != nil {
		return
	}

	typeStr = m.Type()
	if msgID, ok := TypeToCodeDict[typeStr]; ok {
		headerSlice := make([]byte, 2) // 这里需要实际的头部数据
		headerLen = len(headerSlice)
		bodyLen = len(bodySlice)
		out = make([]byte, 2+headerLen+bodyLen)
		binary.LittleEndian.PutUint16(out[:2], uint16(headerLen))
		copy(out[2:2+headerLen], headerSlice)
		copy(out[2+headerLen:], bodySlice)
	} else {
		err = errors.New("TypeToCodeDict not found for type " + typeStr)
	}
	return
}

func AddHead(seq int32, cmd int32, bodySlice []byte) ([]byte, error) {
	headerSlice := make([]byte, 2) // 这里需要实际的头部数据
	headerLenSlice := make([]byte, 2)
	binary.LittleEndian.PutUint16(headerLenSlice, uint16(len(headerSlice)))
	headerLen := len(headerSlice)
	bodyLen := len(bodySlice)
	out := make([]byte, 2+headerLen+bodyLen)
	copy(out[0:2], headerLenSlice)
	copy(out[2:2+headerLen], headerSlice)
	copy(out[2+headerLen:], bodySlice)
	return out, nil
}