package codec

import (
	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet/v2"
)

// Protobuf 实现了gnet.Codec接口，用于实现基于Google protocol buffer解码
type Protobuf struct {
}

func (cp Protobuf) Encode(c gnet.Conn, msg proto.Message) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (cp Protobuf) Decode(c gnet.Conn, msg proto.Message) error {
	var data []byte
	_, err := c.Read(data)

	if err != nil {
		return err
	}

	err = proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	return nil
}
