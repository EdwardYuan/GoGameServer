package codec

import (
	"GoGameServer/src/lib"
	"github.com/panjf2000/gnet"
	"unsafe"
)

//MsgCode实现gnet的Codec接口
type MsgCodec struct {
	//Head   lib.MessageHead
	Head   ServerMessageHead
	Offset uint8
	Data   []byte
}

// Encode encodes frames upon server responses into TCP stream.
func (mc MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

// Decode decodes frames from TCP stream via specific implementation.
// 读取一个完整的消息包；处理组包问题
func (mc MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	var (
		in  inBuffer
		err error
	)
	headlen := int(unsafe.Sizeof(ServerMessageHead{}))
	in = c.Read()
	//size, in := c.ReadN(headlen)
	lib.SugarLogger.Infof("read buffer length %d", headlen)
	buf, err := in.readN(headlen) //in.readN(MessageHeadLength)
	if err != nil {
		return nil, err
	}
	head := new(ServerMessageHead)
	head.Decode(buf)
	// TODO 校验包头
	if ok, err := head.Check(); !ok {
		lib.LogIfError(err, "decode message head error")
	}
	// 读取包头完成
	//idx := unsafe.Sizeof(head)

	c.ShiftN(headlen)
	lib.SugarLogger.Infof("size is %d", head)
	//size, offset := c.ReadN(1)
	//c.ShiftN(size)
	//lib.Log(zap.InfoLevel, string(offset), nil)

	//size, data := c.ReadN(head.DataLength)
	data, err := in.read(headlen+1, headlen+1+head.DataLength)
	in = append(in, data...)
	if err != nil {
		return nil, err
	}
	// TODO 校验包体
	// 返回的是一个完整的消息体
	c.ShiftN(MessageHeadLength + head.DataLength)
	return data, err
}
