// 该代码文件是一个Go语言文件，包含了一个实现gnet的Codec接口的消息编解码器(MsgCodec)以及相关的编解码函数。
// 该文件中包括了对消息的编码(Encode)和解码(Decode)函数实现，以及对消息头(ServerMessageHead)的处理和校验。
// 最终的代码实现了消息的编码和解码过程，确保可以在TCP流中正确地处理消息。

package codec

import (
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"github.com/panjf2000/gnet"
	"google.golang.org/protobuf/proto"
)

// MsgCodec 实现gnet的Codec接口
type MsgCodec struct {
	Head   ServerMessageHead
	Offset uint32
	Data   []byte
}

func EncodeMessage(msg *pb.ProtoInternal) (out []byte, err error) {
	return
}

// 该函数用于解码数据，首先读取消息头部信息，然后根据头部信息读取消息体并生成ProtoInternal结构体返回。

func DecodeData(buf []byte) (msg *pb.ProtoInternal, err error) {
	var (
		in      inBuffer
		readBuf inBuffer
	)
	in = buf
	head := new(ServerMessageHead)
	// todo check offset
	readBuf, err = in.readN(MessageHeadLength)
	head.Decode(readBuf)
	if ok, err := head.Check(); !ok || err != nil {
		if lib.LogErrorAndReturn(err, "Decode head error") {
			return nil, err
		}
	}
	in.ShiftN(MessageHeadLength)
	body, err := in.readN(head.DataLength)
	outMsg := &pb.ProtoInternal{
		Cmd:       int32(head.Cmd),
		Dst:       "",
		SessionId: 0,
		Data:      body,
	}
	msg = outMsg
	return
}

// Encode encodes frames upon server responses into TCP stream.
func (mc MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	msg := &pb.ProtoInternal{}
	err := proto.Unmarshal(buf, msg)
	if lib.LogErrorAndReturn(err, "") {
		return nil, err
	}
	return EncodeMessage(msg)
}

// Decode decodes frames from TCP stream via specific implementation.
// 读取一个完整的消息包；处理组包问题
func (mc MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
    var (
        in   inBuffer
        err  error
        size int
        out  []byte
    )
    head := new(ServerMessageHead)
    if mc.Offset < MessageHeadLength {
        size, in = c.ReadN(MessageHeadLength)
        mc.Offset = uint32(size)
        lib.SugarLogger.Debugf("read buffer length %d", MessageHeadLength)
        buf, err := in.readN(MessageHeadLength)
        if err != nil {
            return nil, err
        }
        head.Decode(buf)
        // TODO 校验包头
        if ok, err := head.Check(); !ok {
            lib.LogIfError(err, "decode message head error")
            // 丢弃
            return nil, err
        }
        // 读取包头完成
        c.ShiftN(MessageHeadLength)
        lib.SugarLogger.Debugf("size is %d", head)
    }
    if mc.Offset < uint32(MessageHeadLength+1+head.DataLength) {
        data, err := in.read(MessageHeadLength+1, MessageHeadLength+1+head.DataLength)
        if lib.LogErrorAndReturn(err, "decode message error") {
            return nil, err
        }
        outMsg := &pb.ProtoInternal{
            Cmd:       int32(head.Cmd),
            SessionId: 0,
            Data:      data,
        }
        out, err = proto.Marshal(outMsg)
        // TODO 校验包体
        // 返回的是一个完整的消息体
        c.ShiftN(MessageHeadLength + head.DataLength)
    }
    return out, err
}
