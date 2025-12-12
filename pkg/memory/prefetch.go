package memory

import (
	"unsafe"
)

//go:noescape
//go:linkname prefetch runtime.prefetch
func prefetch(addr unsafe.Pointer)

func Prefetch(ptr unsafe.Pointer) {
	_ = *(*int)(ptr)
}

func PrefetchRead(data []byte, offset int) {
	if offset < len(data) {
		Prefetch(unsafe.Pointer(&data[offset]))
	}
}

func PrefetchWrite(data []byte, offset int) {
	if offset < len(data) {
		Prefetch(unsafe.Pointer(&data[offset]))
	}
}

func PrefetchStride(data []byte, start, stride, count int) {
	for i := 0; i < count; i++ {
		offset := start + (i * stride)
		if offset < len(data) {
			Prefetch(unsafe.Pointer(&data[offset]))
		}
	}
}
