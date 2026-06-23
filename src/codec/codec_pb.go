package codec

import (
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

// Protobuf 实现了gnet.Codec接口，用于实现基于Google protocol buffer解码
type Protobuf struct {
}

func (cp Protobuf) Encode(c gnet.Conn, msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func (cp Protobuf) Decode(c gnet.Conn, msg proto.Message) error {
	data := make([]byte, c.InboundBuffered())
	if len(data) == 0 {
		return nil
	}
	if _, err := c.Read(data); err != nil {
		return err
	}
	return proto.Unmarshal(data, msg)
}
