// codec_msg.go 提供 ProtoInternal 专用的自定义消息体编码。
//
// 该文件包含两个层次：MsgSerializer 只负责 ProtoInternal 的 Body 编码；
// MsgCodec 是兼容 gnet Codec 风格的薄封装，实际拆包仍复用 FrameCodec。
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

// MsgCodec 是兼容 gnet Codec 风格的消息编解码器。
//
// 新的服务间拆包逻辑集中在 FrameCodec，这个类型只保留给旧调用点过渡使用。
type MsgCodec struct {
	Head   ServerMessageHead
	Offset uint32
	Data   []byte
}

// MsgSerializer 使用固定字段布局序列化 pb.ProtoInternal。
type MsgSerializer struct{}

const (
	// msgCmdLength 是自定义消息体中 Cmd 字段的字节数。
	msgCmdLength = 4

	// msgSessionIDLength 是自定义消息体中 SessionId 字段的字节数。
	msgSessionIDLength = 8

	// msgDstLengthLength 是 Dst 字符串长度字段的字节数。
	msgDstLengthLength = 2

	// msgDataLengthLength 是 Data 字节数组长度字段的字节数。
	msgDataLengthLength = 4

	// msgMinLength 是没有 Dst 和 Data 内容时的最小消息体长度。
	msgMinLength = msgCmdLength + msgSessionIDLength + msgDstLengthLength + msgDataLengthLength

	// msgMaxDstLength 是 Dst 字段可编码的最大字节数。
	msgMaxDstLength = 0xffff
)

// Name 返回自定义 ProtoInternal 编码协议名称。
func (MsgSerializer) Name() string {
	return string(CodecSchemeMsg)
}

// Marshal 将 pb.ProtoInternal 编码为自定义二进制消息体。
//
// 布局为 Cmd、SessionId、Dst 长度、Dst 内容、Data 长度、Data 内容。
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

// Unmarshal 将自定义二进制消息体解析到 pb.ProtoInternal。
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

// EncodeMessage 将内部服务消息序列化并封装为完整服务间消息帧。
//
// 该函数会校验 Cmd 是否能安全写入一字节帧头，避免静默截断。
func EncodeMessage(msg *pb.ProtoInternal) (out []byte, err error) {
	if msg == nil {
		return nil, errors.New("cannot encode nil ProtoInternal message")
	}
	if msg.Cmd < 0 || msg.Cmd > 0xff {
		return nil, fmt.Errorf("message cmd %d exceeds uint8 range", msg.Cmd)
	}
	return DefaultFrameCodec().Encode(uint8(msg.Cmd), 0, msg)
}

// DecodeData 从完整帧字节中解析出兼容旧接口的 pb.ProtoInternal。
//
// 该函数只填充帧头中的 Cmd 和原始 Body；Body 的 protobuf 反序列化由调用方完成。
func DecodeData(buf []byte) (msg *pb.ProtoInternal, err error) {
	frame, _, err := DecodeFrame(buf)
	if err != nil {
		return nil, err
	}
	return &pb.ProtoInternal{
		Cmd:       int32(frame.Head.Cmd),
		Dst:       "",
		SessionId: 0,
		Data:      frame.Body,
	}, nil
}

// decodeFrameBody 从 gnet 连接中读取一个完整帧并返回原始消息体。
func decodeFrameBody(c gnet.Conn) ([]byte, error) {
	frame, err := DefaultFrameCodec().Decode(c)
	if err != nil {
		return nil, err
	}
	return frame.Body, nil
}

// Encode 将 protobuf 格式的 ProtoInternal 包装为完整服务间消息帧。
func (mc MsgCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	msg := &pb.ProtoInternal{}
	err := proto.Unmarshal(buf, msg)
	if lib.LogErrorAndReturn(err, "") {
		return nil, err
	}
	return EncodeMessage(msg)
}

// Decode 从 TCP 流中读取一个完整服务间消息帧，并返回帧 Body。
func (mc MsgCodec) Decode(c gnet.Conn) ([]byte, error) {
	return decodeFrameBody(c)
}
