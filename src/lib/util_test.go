package lib

import (
	"testing"
	"unsafe"
)

type testStruct0 struct {
	name   string
	id     byte
	number uint32
	email  string
}

type testStruct1 struct {
	id      uint64
	brand   string
	carType uint16
	color   uint64
	height  int
	length  int
	width   int
}

type testStruct2 struct {
	cars   map[uint64]*testStruct1
	people []testStruct0
}

func BenchmarkSizeStruct(b *testing.B) {
	//people := make([]testStruct0, b.N)
	//cars := make(map[uint64]*testStruct1, b.N)
	//testStruct2{
	//	cars:   cars,
	//	people: people,
	//}
	for i := 0; i < b.N; i++ {
		SizeStruct(testStruct2{})
	}
}

func BenchmarkSysSizeof(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = unsafe.Sizeof(testStruct2{})
	}
}
