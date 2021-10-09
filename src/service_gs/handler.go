package service_gs

import "GoGameServer/src/lib"

type handler interface {
	Unmarshal(buf []byte) lib.Message
	Marshal(m lib.Message) []byte
}
