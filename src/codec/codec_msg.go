package codec

import (
	"encoding/binary"
	"errors"
	"fmt"

	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	gnet "github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

// MsgCodec 实现gnet的Codec接口
type MsgCodec struct {
	Head   ServerMessageHead
	Offset uint32
	Data   []byte
}

type MsgSerializer struct{}

const (
	msgCmdLength        = 4
	msgSessionIDLength  = 8
	msgDstLengthLength  = 2
	msgDataLengthLength = 4
	msgMinLength        = msgCmdLength + msgSessionIDLength + msgDstLengthLength + msgDataLengthLength
	msgMaxDstLength     = 0xffff
)

func (MsgSerializer) Name() string {
	return string(CodecSchemeMsg)
}

func (MsgSerializer) Marshal(msg proto.Message) ([]byte, error) {
	internal, ok := msg.(*pb.ProtoInternal)
	if !ok {
		return nil, fmt.Errorf("codec_msg only supports *pb.ProtoInternal, got %T", msg)
	}
	dst := []byte(internal.Dst)
	data := internal.Data
	if len(dst) > msgMaxDstLength {
		return nil, fmt.Errorf("dst length %d exceeds max length %d", len(dst), msgMaxDstLength)
	}
	outLen := msgMinLength + len(dst) + len(data)
	out := make([]byte, outLen)
	offset := 0
	binary.LittleEndian.PutUint32(out[offset:offset+msgCmdLength], uint32(internal.Cmd))
	offset += msgCmdLength
	binary.LittleEndian.PutUint64(out[offset:offset+msgSessionIDLength], internal.SessionId)
	offset += msgSessionIDLength
	binary.LittleEndian.PutUint16(out[offset:offset+msgDstLengthLength], uint16(len(dst)))
	offset += msgDstLengthLength
	copy(out[offset:offset+len(dst)], dst)
	offset += len(dst)
	binary.LittleEndian.PutUint32(out[offset:offset+msgDataLengthLength], uint32(len(data)))
	offset += msgDataLengthLength
	copy(out[offset:], data)
	return out, nil
}

func (MsgSerializer) Unmarshal(data []byte, msg proto.Message) error {
	internal, ok := msg.(*pb.ProtoInternal)
	if !ok {
		return fmt.Errorf("codec_msg only supports *pb.ProtoInternal, got %T", msg)
	}
	if len(data) < msgMinLength {
		return errors.New("codec_msg data is too short")
	}
	offset := 0
	internal.Cmd = int32(binary.LittleEndian.Uint32(data[offset : offset+msgCmdLength]))
	offset += msgCmdLength
	internal.SessionId = binary.LittleEndian.Uint64(data[offset : offset+msgSessionIDLength])
	offset += msgSessionIDLength
	dstLen := int(binary.LittleEndian.Uint16(data[offset : offset+msgDstLengthLength]))
	offset += msgDstLengthLength
	if len(data) < offset+dstLen+msgDataLengthLength {
		return errors.New("codec_msg dst length exceeds buffer")
	}
	internal.Dst = string(data[offset : offset+dstLen])
	offset += dstLen
	bodyLen := int(binary.LittleEndian.Uint32(data[offset : offset+msgDataLengthLength]))
	offset += msgDataLengthLength
	if len(data) < offset+bodyLen {
		return errors.New("codec_msg body length exceeds buffer")
	}
	internal.Data = append(internal.Data[:0], data[offset:offset+bodyLen]...)
	if len(data) != offset+bodyLen {
		return fmt.Errorf("codec_msg has %d trailing bytes", len(data)-(offset+bodyLen))
	}
	return nil
}

func EncodeMessage(msg *pb.ProtoInternal) (out []byte, err error) {
	body, err := Marshal(msg)
	if err != nil {
		return nil, err
	}
	return EncodeFrame(uint8(msg.Cmd), 0, body)
}

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

	// buf := c.Read()    // TODO fix with gnet v2
	var buf []byte // tmp
	msg, err := DecodeData(buf)
	lib.LogErrorAndReturn(err, "")
	return msg.Data, err
	/*
		var (
			in   inBuffer
			err  error
			size int
			out  []byte
		)
		head := new(ServerMessageHead)
		if mc.Offset < MessageHeadLength {
			size, in = c.ReadN(MessageHeadLength)
			//in = c.Read()
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
			//in = append(in, data...)
			out, err = proto.Marshal(outMsg)
			// TODO 校验包体
			// 返回的是一个完整的消息体
			c.ShiftN(MessageHeadLength + head.DataLength)
		}
		return out, err
	*/
}
