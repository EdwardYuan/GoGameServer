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
	if len(dst) > 0xffff {
		return nil, fmt.Errorf("dst length %d exceeds max length %d", len(dst), 0xffff)
	}
	outLen := 4 + 8 + 2 + len(dst) + 4 + len(data)
	out := make([]byte, outLen)
	offset := 0
	binary.LittleEndian.PutUint32(out[offset:offset+4], uint32(internal.Cmd))
	offset += 4
	binary.LittleEndian.PutUint64(out[offset:offset+8], internal.SessionId)
	offset += 8
	binary.LittleEndian.PutUint16(out[offset:offset+2], uint16(len(dst)))
	offset += 2
	copy(out[offset:offset+len(dst)], dst)
	offset += len(dst)
	binary.LittleEndian.PutUint32(out[offset:offset+4], uint32(len(data)))
	offset += 4
	copy(out[offset:], data)
	return out, nil
}

func (MsgSerializer) Unmarshal(data []byte, msg proto.Message) error {
	internal, ok := msg.(*pb.ProtoInternal)
	if !ok {
		return fmt.Errorf("codec_msg only supports *pb.ProtoInternal, got %T", msg)
	}
	if len(data) < 18 {
		return errors.New("codec_msg data is too short")
	}
	offset := 0
	internal.Cmd = int32(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	internal.SessionId = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	dstLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	if len(data) < offset+dstLen+4 {
		return errors.New("codec_msg dst length exceeds buffer")
	}
	internal.Dst = string(data[offset : offset+dstLen])
	offset += dstLen
	bodyLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
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
