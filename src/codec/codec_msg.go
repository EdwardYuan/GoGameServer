package codec

import (
	"GoGameServer/src/lib"

	"github.com/panjf2000/gnet"
	"go.uber.org/zap"
)

//MsgCode实现gnet的Codec接口
type MsgCodec struct {
	Head   lib.MessageHead
	Offset uint8
	Data   []byte
}

// Encode encodes frames upon server responses into TCP stream.
func (mc *MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

// Decode decodes frames from TCP stream via specific implementation.
// 读取一个完整的消息包；处理组包问题
func (mc *MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	var (
		in  inBuffer
		err error
	)
	in = c.Read()
	buf, err := in.readN(MessageHeadLength)
	if err != nil {
		return nil, err
	}
	head := new(ServerMessageHead)
	head.Decode(buf)
	// TODO 校验包头
	err = head.Check()
	if err != nil {
		lib.Log(zap.ErrorLevel, "decode message head error", err)
		return nil, err
	}
	data, err := in.readN(head.DataLength)
	in = append(in, data...)
	if err != nil {
		return nil, err
	}
	// TODO 校验包体
	// 返回的是一个完整的消息体
	return in, err
}
