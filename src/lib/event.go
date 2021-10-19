package lib

import (
	"reflect"
	"strings"
)

type Event interface{}

type SeqEvent struct {
	Seq   int
	Event Event
}

func Name(e Event) string {
	typeName := reflect.TypeOf(e).String()
	names := strings.Split(typeName, ".")
	return names[len(names)-1]
}
