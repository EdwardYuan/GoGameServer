// codec_legend.go 保留早期内网协议编解码器的兼容占位实现。
//
// 当前服务间通信主路径已经收敛到 FrameCodec；该文件中的 CodecLegend
// 仅用于兼容旧接口或后续重新接入早期协议。
package codec

import (
	gnet "github.com/panjf2000/gnet/v2"
)

// CodecLegend 表示早期内网协议的 gnet Codec 兼容类型。
type CodecLegend struct {
}

// Encode 当前直接透传调用方传入的字节。
func (cl CodecLegend) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

// Decode 当前未实现实际解码逻辑，保留旧协议接入点。
func (cl CodecLegend) Decode(c gnet.Conn) ([]byte, error) {
	//in := c.Read()
	//if unsafe.Sizeof(in) > ServerMaxReceiveLength {
	//	c.ResetBuffer()
	//	return nil, nil
	//}
	//
	//return c.Read(), nil
	return nil, nil
}
