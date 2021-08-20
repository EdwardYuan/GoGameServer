package lib

import (
	"errors"

	"github.com/panjf2000/gnet"
)

//MsgCode实现gnet的Codec接口
type MsgCodec struct {
}

// Encode encodes frames upon server responses into TCP stream.
func (mc *MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

// Decode decodes frames from TCP stream via specific implementation.
// 读取一个完整的消息包；处理组包问题
func (mc *MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	size, buf := c.ReadN(ReadMessageInitLength)
	if size == 0 {
		return nil, errors.New("")
	}
	c.ShiftN(size)
	head := NewMessageHead()
	if size < int(head.HeaderLength) {
		// Continue Read
		size, buf1 := c.ReadN(int(head.HeaderLength) - size)
		c.ShiftN(size)
		if size < int(head.HeaderLength) {
			return nil, nil
		}
		buf = append(buf, buf1...)
	}
	head.Decode(buf)
	if (head.BodyLength == 0) || (head.BodyLength > MaxMessageBodySize) {
		return nil, errors.New("head.bodylength is zero or too large")
	}
	// 校验包头完成，读取包体
	bodySize, data := c.ReadN(int(head.BodyLength))
	c.ShiftN(bodySize)
	if bodySize == 0 {
		return nil, nil
	}
	buf = append(buf, data...)
	return buf, nil
}
