//NOTE: the mechanism in this file is kind of a hack, please BE AWARE!
package protocol

import (
	"GoGameServer/src/lib"
	"errors"
	"reflect"
	"runtime/debug"
)

var ErrGenericDecodeError = errors.New("generic decode error")

func ParseProtobufEvent(bs []byte) (seq int, e lib.Event, typeStr string, headerLen, bodyLen int, err error) {
	// if !config.FlagDev {
	defer func() {
		if r := recover(); r != nil {
			lib.SugarLogger.Errorf("recover from proto event parse error event_name %s recover_result %v debug_stack %s",
				lib.Name(e), r, string(debug.Stack()))
			switch r := r.(type) {
			case string:
				err = errors.New(r)
			case error:
				err = r
			default:
				err = ErrGenericDecodeError
			}
		}
	}()
	// }

	m, headerLen, bodyLen, err := Decode(bs)
	if err != nil {
		return
	}
	seq = int(m.ID)
	typeStr = m.Type()

	methodValue, ok := decMethods[typeStr]
	if ok {
		// 手动从protobuf转到event
		rv := methodValue.Call([]reflect.Value{reflect.ValueOf(m.Body)})
		e = rv[0].Interface()
	} else {
		// 直接解码
		e = m.Body.(lib.Event)
	}

	return
}

type PBDecoder interface {
	PkgName() string
}

//type DecBattle struct{}
type DecChat struct{}
type DecQuery struct{}
type DecLogin struct{}
type DecShared struct{}
type DecField struct{}
type DecAction struct{}
type DecFriend struct{}

//func (d DecBattle) PkgName() string {
//	return "battle"
//}
func (d DecChat) PkgName() string {
	return "chat"
}
func (d DecQuery) PkgName() string {
	return "query"
}
func (d DecLogin) PkgName() string {
	return "login"
}
func (d DecShared) PkgName() string {
	return "shared"
}
func (d DecField) PkgName() string {
	return "field"
}
func (d DecAction) PkgName() string {
	return "action"
}
func (d DecFriend) PkgName() string {
	return "friend"
}

var (
	//	decBattle DecBattle
	decChat   DecChat
	decQuery  DecQuery
	decLogin  DecLogin
	decShared DecShared
	decField  DecField
	decAction DecAction
	decFriend DecFriend
)

var decMethods map[string]reflect.Value = map[string]reflect.Value{}

func init() {
	//	registerDecoder(decBattle)
	registerDecoder(decChat)
	registerDecoder(decQuery)
	registerDecoder(decLogin)
	registerDecoder(decShared)
	registerDecoder(decField)
	registerDecoder(decAction)
	registerDecoder(decFriend)
}

func registerDecoder(dec PBDecoder) {
	v := reflect.ValueOf(dec)
	t := v.Type()
	n := t.NumMethod()
	for i := 0; i < n; i++ {
		method := t.Method(i)
		if method.Name == "PkgName" { // this is in interface{}
			continue
		}
		name := dec.PkgName() + "." + method.Name
		if _, ok := decMethods[name]; ok {
			panic("decode method " + name + " already registered")
		}
		decMethods[name] = v.Method(i)
	}
}

/*
func (DecShared) Ping(pm proto.Message) *gmevt.Ping {
	return new(gmevt.Ping)
}
*/
