package protocol

import "GoGameServer/src/lib"

// ProtocolFactory
type Factory struct {
}

// Session
type Session struct {
}

func (s *Session) Decode(msg lib.Message) {

}

func (s *Session) Encode(buf []byte, len int) {

}

func (f *Factory) ReadMessage() {

}
