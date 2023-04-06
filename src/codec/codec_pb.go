package codec

import (
	gnet "github.com/panjf2000/gnet/v2"
)

// CodecProtobuf protocol buffer解码
type CodecProtobuf struct {
}

func (cp CodecProtobuf) Encode(c gnet.Conn, buf []byte) ([]byte, error) {

	return buf, nil
}

func (cp CodecProtobuf) Decode(c gnet.Conn) ([]byte, error) {
	//in := c.Read()
	//idx := len(in)
	//c.ShiftN(idx)
	//return in, nil
	return nil, nil
}
