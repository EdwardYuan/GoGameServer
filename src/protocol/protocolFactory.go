package protocol

import "GoGameServer/src/MsgHandler"

// ProtocolFactory
type Factory struct {
}

// Session
type Session struct {
}

func (s *Session) Decode(msg MsgHandler.Message) {

}

func (s *Session) Encode(buf []byte, len int) {

}

func (f *Factory) ReadMessage() {

}
