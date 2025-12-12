package memory

import (
	"fmt"
	"reflect"
	"unsafe"
)

const CacheLineSize = 64

type AlignedBuffer struct {
	data []byte
	ptr  unsafe.Pointer
}

func AllocAligned(size, alignment int) (*AlignedBuffer, error) {
	if alignment <= 0 || (alignment&(alignment-1)) != 0 {
		return nil, fmt.Errorf("alignment must be power of 2")
	}

	buf := make([]byte, size+alignment)
	ptr := unsafe.Pointer(&buf[0])
	aligned := unsafe.Pointer((uintptr(ptr) + uintptr(alignment) - 1) &^ (uintptr(alignment) - 1))

	return &AlignedBuffer{
		data: buf,
		ptr:  aligned,
	}, nil
}

func (a *AlignedBuffer) Ptr() unsafe.Pointer {
	return a.ptr
}

func (a *AlignedBuffer) Slice(size int) []byte {
	var s []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Data = uintptr(a.ptr)
	sh.Len = size
	sh.Cap = size
	return s
}

func AllocCacheLine() (*AlignedBuffer, error) {
	return AllocAligned(CacheLineSize, CacheLineSize)
}

func AllocCacheLines(count int) (*AlignedBuffer, error) {
	return AllocAligned(count*CacheLineSize, CacheLineSize)
}

func IsAligned(ptr unsafe.Pointer, alignment int) bool {
	return uintptr(ptr)%uintptr(alignment) == 0
}

func Pad64(size int) int {
	remainder := size % CacheLineSize
	if remainder == 0 {
		return size
	}
	return size + (CacheLineSize - remainder)
}
