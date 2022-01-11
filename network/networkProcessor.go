package network

import "GoGameServer/src/pb"

type NetworkProcessor struct {
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
