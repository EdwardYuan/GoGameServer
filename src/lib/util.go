package lib

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"sync/atomic"
	"unsafe"
)

type IPVersion int

const (
	IPv4 IPVersion = iota
	IPv6
)

// AtomBool 原子Bool
type AtomBool struct {
	flag int32
}

func (b *AtomBool) Set(value bool) {
	var flag int32 = 0
	if value {
		flag = 1
	}
	atomic.StoreInt32(&(b.flag), flag)
}

func (b *AtomBool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) == 1
}

func SizeStruct(data interface{}) int {
	return sizeof(reflect.ValueOf(data))
}

// sizeof 计算结构体实际占用空间大小而不是对齐大小  注意!!! 性能太差 能不用尽量不用
func sizeof(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Map:
		sum := 0
		keys := v.MapKeys()
		for i := 0; i < len(keys); i++ {
			mapkey := keys[i]
			s := sizeof(mapkey)
			if s < 0 {
				return -1
			}
			sum += s
			s = sizeof(v.MapIndex(mapkey))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum
	case reflect.Slice, reflect.Array:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.String:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Ptr, reflect.Interface:
		p := (*[]byte)(unsafe.Pointer(v.Pointer()))
		if p == nil {
			return 0
		}
		return sizeof(v.Elem())
	case reflect.Struct:
		sum := 0
		for i, n := 0, v.NumField(); i < n; i++ {
			s := sizeof(v.Field(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Int:
		return int(v.Type().Size())

	default:
		fmt.Println("t.Kind() no found:", v.Kind())
	}

	return -1
}

func GetLocalIP(ver IPVersion) string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		switch ver {
		case IPv4:
			if ip4 := addr.To4(); ip4 != nil {
				return ip4.String()
			}
		case IPv6:
			if ip6 := addr.To16(); ip6 != nil {
				return ip6.String()
			}
		}
	}
	return ""
}
