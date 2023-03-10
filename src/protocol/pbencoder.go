// NOTE: the mechanism in this file is kind of a hack, please BE AWARE!
package protocol

import (
	"GoGameServer/src/lib"
	"errors"
	"reflect"
	"runtime/debug"
	"strings"

	"google.golang.org/protobuf/proto"
)

var ErrUnknown = errors.New("unknown error")

func GenerateEventProtobuf(seq int, e lib.Event) (bs []byte, name string, headerLen, bodyLen int, err error) {
	// if !config.FlagDev {
	defer func() {
		if r := recover(); r != nil {
			lib.SugarLogger.Error("recover from proto event generate error", "event_name", lib.Name(e),
				"recover_result", r, "debug_stack", string(debug.Stack()))
			switch r := r.(type) {
			case string:
				err = errors.New(r)
			case error:
				err = r
			default:
				err = ErrUnknown
			}
		}
	}()
	// }

	out := MessageProtoFromEvent(seq, e)
	return out.Encode()
}

func MessageProtoFromEvent(seq int, e lib.Event) *Message {
	var m proto.Message

	//NOTE: assume concrete event is a pointer
	t := reflect.ValueOf(e).Elem().Type()
	pkgSlice := strings.Split(t.PkgPath(), "/")
	pkgName := pkgSlice[len(pkgSlice)-1]
	eventName := pkgName + "." + t.Name()

	//println("GenerateEventProtobuf:", eventName)
	//	methodValue, ok := encMethods[eventName]
	//	if !ok {
	//		panic(: no registered method for event " + eventName)
	//	}

	methodValue, ok := encMethods[eventName]
	if ok {
		// 手动从event转到protobuf
		arg := reflect.ValueOf(e)
		rv := methodValue.Call([]reflect.Value{arg})
		m, ok = rv[0].Interface().(proto.Message)
		if !ok {
			panic("Cannot type assert return value to proto.Message")
		}
		if m == nil {
			panic("nil message probably should not reach here")
		}
	} else {
		// 直接编码
		m = e.(proto.Message)
	}
	return NewMessageProto(int32(seq), m)
}

type PBEncoder interface {
	PkgName() string
}

// EncGame type EncBattle struct{}
type EncGame struct{}

//	func (e EncBattle) PkgName() string {
//		return "btevt"
//	}
func (e EncGame) PkgName() string {
	return "gmevt"
}

var (
	//	encBattle EncBattle
	encGame EncGame
)

var encMethods = map[string]reflect.Value{}

func init() {
	//	registerEncoder(encBattle)
	registerEncoder(encGame)
}

func registerEncoder(enc PBEncoder) {
	v := reflect.ValueOf(enc)
	t := v.Type()
	n := t.NumMethod()
	for i := 0; i < n; i++ {
		method := t.Method(i)
		if method.Name == "PkgName" { // this is in interface{}
			continue
		}
		name := enc.PkgName() + "." + method.Name
		if _, ok := encMethods[name]; ok {
			panic("encode method " + name + " already registered")
		}
		encMethods[name] = v.Method(i)
	}
}

// func (EncGame) Pong(e *gmevt.Pong) *shared.Pong {
// return new(shared.Pong)
// }

// func (EncGame) ServerMessage(e *gmevt.ServerMessage) *shared.ServerMessage {
// 	return &shared.ServerMessage{
// 		Content: e.Content,
// 		Code:    e.Code,
// 	}
// }

// func fieldPos(x int) *field.FieldPos {
// 	return &field.FieldPos{X: int32(x)}
// }

// func fieldEquips(equips []gmevt.Equip) []*field.Equip {
// 	rv := []*field.Equip{}
// 	for _, v := range equips {
// 		rv = append(rv, &field.Equip{
// 			Id: int32(v.ID),
// 			//RecastTimes: proto.Int(v.RecastTimes),
// 			FashionID: int32(v.FashionID),
// 		})
// 	}

// 	return rv
// }
