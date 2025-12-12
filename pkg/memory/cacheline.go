package memory

import (
	"unsafe"
)

type CacheLinePadded [CacheLineSize]byte

func (c *CacheLinePadded) Uint32Ptr(offset int) *uint32 {
	return (*uint32)(unsafe.Pointer(&c[offset]))
}

func (c *CacheLinePadded) Uint64Ptr(offset int) *uint64 {
	return (*uint64)(unsafe.Pointer(&c[offset]))
}

func (c *CacheLinePadded) Int32Ptr(offset int) *int32 {
	return (*int32)(unsafe.Pointer(&c[offset]))
}

func (c *CacheLinePadded) Int64Ptr(offset int) *int64 {
	return (*int64)(unsafe.Pointer(&c[offset]))
}

func (c *CacheLinePadded) ByteSlice(start, length int) []byte {
	return c[start : start+length]
}

type CounterLine struct {
	_             [0]CacheLinePadded
	BanCount      uint32
	KickCount     uint32
	ChannelDelete uint32
	RoleDelete    uint32
	WebhookCreate uint32
	PermChange    uint32
	Reserved1     uint32
	Reserved2     uint32
	_             [CacheLineSize - 32]byte
}

func NewCounterLine() *CounterLine {
	buf, _ := AllocCacheLine()
	return (*CounterLine)(buf.Ptr())
}
