package network

import (
	"GoGameServer/src/codec"
	"GoGameServer/src/lib"
	"GoGameServer/src/pb"
	"errors"
	"go.uber.org/zap"
)

type Processor struct {
	session        *Session
	reader         MessageReader
	writer         MessageWriter
	writeChan      chan *pb.ProtoInternal
	readCloseChan  chan int
	writeCloseChan chan int
	logger         *zap.SugaredLogger
}

func NewProcessor(s *Session) *Processor {
	p := &Processor{
		reader:         NewLineBasedMessageReader(s.logger),
		writer:         NewLineBasedMessageWriter(s.logger),
		writeChan:      make(chan *pb.ProtoInternal, lib.MaxMessageCount),
		readCloseChan:  make(chan int),
		writeCloseChan: make(chan int),
		logger:         s.logger,
	}
	p.session = s
	return p
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
	return codec.DecodeData(lr.readBuffer)

	//------------------------------------------------
	lib.LogErrorAndReturn(err, "")
	lr.readOffset = uint32(size)
	head := new(codec.ServerMessageHead)
	if lr.readOffset != codec.MessageHeadLength {
		err = errors.New("")
		lib.LogErrorAndReturn(err, "")
	}
	head.Decode(lr.readBuffer[:codec.MessageHeadLength])
	head.Check()
	size, err = session.conn.Read(lr.readBuffer[:lr.head.DataLength])
	lib.LogErrorAndReturn(err, "")
	msgInternal := &pb.ProtoInternal{
		Cmd:       int32(head.Cmd),
		Dst:       "",
		SessionId: uint64(session.id),
		Data:      lr.readBuffer,
	}
	return msgInternal, err
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

	buf, err := codec.EncodeMessage(message)
	copy(lw.writeBuffer, buf)
	return err
	/*
		var buf []byte
		lw.head.Encode(buf)
		lw.writeBuffer = append(lw.writeBuffer, buf...)
		data, err := proto.Marshal(message)
		lib.LogErrorAndReturn(err, "")
		lw.writeBuffer = append(lw.writeBuffer, data...)

		return err
	*/
}
