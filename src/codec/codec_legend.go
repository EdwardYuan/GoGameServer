package codec

import "github.com/panjf2000/gnet"

// CodecLegend 内网实现协议
type CodecLegend struct {
}

func (cl CodecLegend) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (cl CodecLegend) Decode(c gnet.Conn) ([]byte, error) {
	return c.Read(), nil
}
