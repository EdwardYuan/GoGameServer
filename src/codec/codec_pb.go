package codec

import (
	"errors"
	"fmt"

	"github.com/panjf2000/gnet"
	"google.golang.org/protobuf/proto"
)

// CodecProtobuf protocol buffer解码
type CodecProtobuf struct {
}

func (cp CodecProtobuf) Encode(c gnet.Conn, msg interface{}) ([]byte, error) {
	pbMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, errors.New(fmt.Sprintf("failed to encode. msg is not a proto.Message"))
	}

	buf, err := proto.Marshal(pbMsg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to encode. error: %v", err))
	}
	return buf, err
}

func (cp CodecProtobuf) Decode(c gnet.Conn, msg proto.Message) (interface{}, error) {
	buf := c.Read()
	if len(buf) == 0 {
		return nil, nil
	}
	err := proto.Unmarshal(buf, msg)
	if err != nil {
		return nil, err
	}
	// 清空已经读取的数据
	c.ShiftN(len(buf))
	return msg, nil
}

func (cp CodecProtobuf) Release(c gnet.Conn) error {
	c.ResetBuffer()
	return nil
}
