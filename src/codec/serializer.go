package codec

import (
	"errors"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
)

var ErrNilSerializer = errors.New("serializer is nil")

type CodecScheme string

const (
	CodecSchemeProtobuf CodecScheme = "protobuf"
	CodecSchemeMsg      CodecScheme = "codec_msg"
)

type Serializer interface {
	Name() string
	Marshal(proto.Message) ([]byte, error)
	Unmarshal([]byte, proto.Message) error
}

type ProtobufSerializer struct{}

func (ProtobufSerializer) Name() string {
	return "protobuf"
}

func (ProtobufSerializer) Marshal(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func (ProtobufSerializer) Unmarshal(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

var (
	defaultSerializerMu sync.RWMutex
	defaultSerializer   Serializer = ProtobufSerializer{}
)

func DefaultSerializer() Serializer {
	defaultSerializerMu.RLock()
	defer defaultSerializerMu.RUnlock()
	return defaultSerializer
}

func SetDefaultSerializer(serializer Serializer) error {
	if serializer == nil {
		return ErrNilSerializer
	}
	defaultSerializerMu.Lock()
	defaultSerializer = serializer
	defaultSerializerMu.Unlock()
	return nil
}

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

func Marshal(msg proto.Message) ([]byte, error) {
	return DefaultSerializer().Marshal(msg)
}

func Unmarshal(data []byte, msg proto.Message) error {
	return DefaultSerializer().Unmarshal(data, msg)
}
