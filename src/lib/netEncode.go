package lib

import "github.com/panjf2000/gnet"

//MsgCode实现gnet的Codec接口
type MsgCodec struct {
}

// Encode encodes frames upon server responses into TCP stream.
func (mc *MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

// Decode decodes frames from TCP stream via specific implementation.
func (mc *MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	buf := c.Read()
	if len(buf) == 0 {
		return nil, nil
	}
	c.ResetBuffer()
	return buf, nil
}
