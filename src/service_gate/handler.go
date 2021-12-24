package service_gate

//import "google.golang.org/protobuf/proto"
import "github.com/gogo/protobuf/proto"

type MessageHandler struct {
}

func (h *MessageHandler) Check(message proto.Message) bool {
	return true
}
