package MsgHandler

type Message struct {
	Head    MessageHead
	Data    []byte
	DataLen int32
}

type MessageHead struct {
}
