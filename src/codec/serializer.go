// serializer.go 定义消息体序列化抽象和默认序列化策略。
//
// FrameCodec 只负责帧边界，具体 Body 如何 marshal/unmarshal 由 Serializer 决定。
package codec

import (
	"errors"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
)

// ErrNilSerializer 表示调用方试图设置空的默认序列化器。
var ErrNilSerializer = errors.New("serializer is nil")

// CodecScheme 表示内置序列化协议的名称。
type CodecScheme string

const (
	// CodecSchemeProtobuf 表示使用标准 protobuf 序列化消息体。
	CodecSchemeProtobuf CodecScheme = "protobuf"

	// CodecSchemeMsg 表示使用自定义 ProtoInternal 二进制布局序列化消息体。
	CodecSchemeMsg CodecScheme = "codec_msg"
)

// Serializer 抽象服务间消息体的序列化和反序列化能力。
type Serializer interface {
	Name() string
	Marshal(proto.Message) ([]byte, error)
	Unmarshal([]byte, proto.Message) error
}

// ProtobufSerializer 使用 google protobuf 作为消息体编码格式。
type ProtobufSerializer struct{}

// Name 返回 protobuf 序列化器的协议名称。
func (ProtobufSerializer) Name() string {
	return "protobuf"
}

// Marshal 使用 protobuf 将消息编码为字节。
func (ProtobufSerializer) Marshal(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// Unmarshal 使用 protobuf 将字节解析到目标消息。
func (ProtobufSerializer) Unmarshal(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

var (
	defaultSerializerMu sync.RWMutex
	defaultSerializer   Serializer = ProtobufSerializer{}
)

// DefaultSerializer 返回当前全局默认消息体序列化器。
func DefaultSerializer() Serializer {
	defaultSerializerMu.RLock()
	defer defaultSerializerMu.RUnlock()
	return defaultSerializer
}

// SetDefaultSerializer 设置全局默认消息体序列化器。
func SetDefaultSerializer(serializer Serializer) error {
	if serializer == nil {
		return ErrNilSerializer
	}
	defaultSerializerMu.Lock()
	defaultSerializer = serializer
	defaultSerializerMu.Unlock()
	return nil
}

// SetDefaultCodecScheme 按内置协议名称切换全局默认序列化器。
func SetDefaultCodecScheme(scheme CodecScheme) error {
	switch scheme {
	case CodecSchemeProtobuf:
		return SetDefaultSerializer(ProtobufSerializer{})
	case CodecSchemeMsg:
		return SetDefaultSerializer(MsgSerializer{})
	default:
		return fmt.Errorf("unsupported codec scheme %q", scheme)
	}
}

// Marshal 使用当前默认序列化器编码消息体。
func Marshal(msg proto.Message) ([]byte, error) {
	return DefaultSerializer().Marshal(msg)
}

// Unmarshal 使用当前默认序列化器解析消息体。
func Unmarshal(data []byte, msg proto.Message) error {
	return DefaultSerializer().Unmarshal(data, msg)
}
