// codec_pb.go 提供直接读取连接缓冲区的 protobuf 编解码兼容实现。
//
// 该实现不处理服务间固定帧头，只适用于调用方已经明确需要纯 protobuf 流的场景。
package codec

import (
	"github.com/panjf2000/gnet/v2"
	"google.golang.org/protobuf/proto"
)

// Protobuf 实现基于 google protobuf 的简单消息体编解码。
type Protobuf struct {
}

// Encode 将 protobuf 消息直接序列化为字节。
func (cp Protobuf) Encode(c gnet.Conn, msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// Decode 从连接当前入站缓冲区读取全部数据并反序列化到目标消息。
func (cp Protobuf) Decode(c gnet.Conn, msg proto.Message) error {
	data := make([]byte, c.InboundBuffered())
	if len(data) == 0 {
		return nil
	}
	if _, err := c.Read(data); err != nil {
		return err
	}
	return proto.Unmarshal(data, msg)
}
