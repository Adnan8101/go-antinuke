package util

import (
	"reflect"
	"unsafe"
)

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func SliceU32(data []byte, offset int) uint32 {
	return *(*uint32)(unsafe.Pointer(&data[offset]))
}

func SliceU64(data []byte, offset int) uint64 {
	return *(*uint64)(unsafe.Pointer(&data[offset]))
}

func SliceI32(data []byte, offset int) int32 {
	return *(*int32)(unsafe.Pointer(&data[offset]))
}

func SliceI64(data []byte, offset int) int64 {
	return *(*int64)(unsafe.Pointer(&data[offset]))
}

func WriteU32(data []byte, offset int, val uint32) {
	*(*uint32)(unsafe.Pointer(&data[offset])) = val
}

func WriteU64(data []byte, offset int, val uint64) {
	*(*uint64)(unsafe.Pointer(&data[offset])) = val
}

func WriteI32(data []byte, offset int, val int32) {
	*(*int32)(unsafe.Pointer(&data[offset])) = val
}

func WriteI64(data []byte, offset int, val int64) {
	*(*int64)(unsafe.Pointer(&data[offset])) = val
}

func FindByte(data []byte, target byte) int {
	for i, b := range data {
		if b == target {
			return i
		}
	}
	return -1
}

func FindByteFrom(data []byte, start int, target byte) int {
	for i := start; i < len(data); i++ {
		if data[i] == target {
			return i
		}
	}
	return -1
}
