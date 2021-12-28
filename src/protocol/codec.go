package protocol

import (
	"GoGameServer/src/lib"
	"encoding/binary"
	"errors"
	"reflect"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	ID   int32
	Body proto.Message
}

func (m *Message) Type() string {
	return string(proto.MessageName(m.Body))
}

func nameToType(name string) (reflect.Type, bool) {
	//msgType := proto.MessageType(name)
	var msgType reflect.Type
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

func init() { /*
		for id, name := range dict.EDict_name {
			if strings.HasPrefix(name, "pb_") {
				name = strings.Replace(name, "_", ".", 2)
				name = strings.Replace(name, ".", "_", 1)
			} else {
				name = strings.Replace(name, "_", ".", 1)
			}
			CodeToTypeDict[id] = name
			TypeToCodeDict[name] = id
		} */
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
	// headerSlice := bs[2 : 2+headerLen]
	bodySlice := bs[2+headerLen:]

	bodyLen = totalLen - 2 - headerLen

	// // header := &packet.PacketHead{}
	// if err = proto.Unmarshal(headerSlice, header); err != nil {
	// 	return
	// }
	// ID := header.GetFId()

	var name string
	// if msgID := header.GetFMsgid(); msgID != 0 {
	// 	name = CodeToTypeDict[msgID]
	// } else {
	// 	err = fmt.Errorf("message id %d not supported in CodeToTypeDict for decoding", msgID)
	// 	return
	// }

	t, ok := nameToType(name)
	if !ok {
		// err = fmt.Errorf("message type %v not supported for decoding", header.FType)
		return
	}
	//TODO 不用反射
	v := reflect.New(t.Elem())
	pm := v.Interface().(proto.Message)
	if err = proto.Unmarshal(bodySlice, pm); err != nil {
		lib.SugarLogger.Warnf("Unmarshal failed name %s", name)
		return
	}

	// msg = &Message{ID, pm}
	return
}

func Encode(m *Message) ([]byte, string, int, int, error) {
	return m.Encode()
}

func (m *Message) Encode() (out []byte, typeStr string, headerLen, bodyLen int, err error) {
	var bodySlice []byte

	bodySlice, err = proto.Marshal(m.Body)
	if err != nil {
		return
	}

	// var header *packet.PacketHead
	// typeStr = m.Type()
	// if msgID, ok := TypeToCodeDict[typeStr]; ok {
	// 	header = &packet.PacketHead{FMsgid: msgID}
	// } else {
	// 	err = fmt.Errorf("TypeToCodeDict not found type %s", typeStr)
	// 	return
	// }
	// // if m.ID != 0 {
	// // 	header.FId = m.ID
	// // }
	// headerSlice, err := proto.Marshal(header)
	// if err != nil {
	// 	return
	// }

	var headerSlice []byte // 删除 编译通过用

	headerLenSlice := make([]byte, 2)
	binary.LittleEndian.PutUint16(headerLenSlice, uint16(len(headerSlice)))
	headerLen = len(headerSlice)
	bodyLen = len(bodySlice)
	out = make([]byte, 2+headerLen+bodyLen)
	copy(out[0:2], headerLenSlice)
	copy(out[2:2+headerLen], headerSlice)
	copy(out[2+headerLen:2+headerLen+bodyLen], bodySlice)
	return
}

func AddHead(seq int32, cmd int32, bodySlice []byte) ([]byte, error) {
	//*************************************************
	// header := &packet.PacketHead{FId: seq, FMsgid: cmd}
	// headerSlice, err := proto.Marshal(header)
	// if err != nil {
	// 	return nil, err
	// }
	///// 删除 编译通过用
	headerSlice := make([]byte, 2)
	//////
	headerLenSlice := make([]byte, 2)
	binary.LittleEndian.PutUint16(headerLenSlice, uint16(len(headerSlice)))
	headerLen := len(headerSlice)
	bodyLen := len(bodySlice)
	out := make([]byte, 2+headerLen+bodyLen)
	copy(out[0:2], headerLenSlice)
	copy(out[2:2+headerLen], headerSlice)
	copy(out[2+headerLen:2+headerLen+bodyLen], bodySlice)
	return out, nil
}
