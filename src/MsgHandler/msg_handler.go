package MsgHandler

type Message struct {
	Head    []byte
	Data    []byte
	DataLen int32
}
