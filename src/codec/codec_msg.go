package codec

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"github.com/panjf2000/gnet"
	"google.golang.org/protobuf/proto"
)

//MsgCodec 实现gnet的Codec接口
type MsgCodec struct {
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

	in = c.Read()
	lib.SugarLogger.Debugf("read buffer length %d", MessageHeadLength)
	buf, err := in.readN(MessageHeadLength)
	if err != nil {
		return nil, err
	}
	head := new(ServerMessageHead)
	head.Decode(buf)
	// TODO 校验包头
	if ok, err := head.Check(); !ok {
		lib.LogIfError(err, "decode message head error")
		// 丢弃
	}
	// 读取包头完成
	c.ShiftN(MessageHeadLength)
	lib.SugarLogger.Debugf("size is %d", head)
	data, err := in.read(MessageHeadLength+1, MessageHeadLength+1+head.DataLength)
	if lib.LogErrorAndReturn(err, "decode message error") {
		return nil, err
	}
	outMsg := &pb.ProtoInternal{
		Cmd:       int32(head.Cmd),
		SessionId: 0,
		Data:      data,
	}
	//in = append(in, data...)
	out, err := proto.Marshal(outMsg)
	// TODO 校验包体
	// 返回的是一个完整的消息体
	c.ShiftN(MessageHeadLength + head.DataLength)
	return out, err
}
