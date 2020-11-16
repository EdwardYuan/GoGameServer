package lib

import "github.com/panjf2000/gnet"

type MsgCodec struct {
}

// Encode encodes frames upon server responses into TCP stream.
func (mc *MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return []byte{}, nil
}

// Decode decodes frames from TCP stream via specific implementation.
func (mc *MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	return nil, nil
}
