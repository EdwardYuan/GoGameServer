package network

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"go.uber.org/zap"
)

type Processor struct {
	session    *Session
	readBuffer []byte
	readOffset int
	network    *Network
}

type MessageReader interface {
	ReadMessage(session *Session) (*pb.ProtoInternal, error)
}

type MessageWriter interface {
	WriteMessage(session *Session, message *pb.ProtoInternal) error
}

type LineBasedMessageReader struct {
	head          *codec.ServerMessageHead
	maxBodyLength uint32
	readBuffer    []byte
	readOffset    uint32
	logger        *zap.SugaredLogger
}

func NewLineBasedMessageReader(logger *zap.SugaredLogger) *LineBasedMessageReader {
	return &LineBasedMessageReader{
		readBuffer: make([]byte, lib.MaxReceiveBufCap),
		logger:     logger,
	}
}

func (lr *LineBasedMessageReader) ReadMessage(session *Session) (*pb.ProtoInternal, error) {
	size, err := session.conn.Read(lr.readBuffer[lr.readOffset:codec.MessageHeadLength])
	lib.LogErrorAndReturn(err, "")
	lr.readOffset = uint32(size)
	return nil, nil
}

type LineBasedMessageWriter struct {
	head          *codec.ServerMessageHead
	maxBodyLength uint32
	writeBuffer   []byte
	logger        *zap.SugaredLogger
}

func NewLineBasedMessageWriter(logger *zap.SugaredLogger) *LineBasedMessageWriter {
	return &LineBasedMessageWriter{
		writeBuffer: make([]byte, lib.MaxReceiveBufCap),
		logger:      logger,
	}
}

func (lw *LineBasedMessageWriter) WriteMessage(session *Session, message *pb.ProtoInternal) error {
	return nil
}
