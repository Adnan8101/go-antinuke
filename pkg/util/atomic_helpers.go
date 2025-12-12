package util

import (
	"sync/atomic"
	"unsafe"
)

func AtomicLoadU32(ptr *uint32) uint32 {
	return atomic.LoadUint32(ptr)
}

func AtomicStoreU32(ptr *uint32, val uint32) {
	atomic.StoreUint32(ptr, val)
}

func AtomicAddU32(ptr *uint32, delta uint32) uint32 {
	return atomic.AddUint32(ptr, delta)
}

func AtomicCASU32(ptr *uint32, old, new uint32) bool {
	return atomic.CompareAndSwapUint32(ptr, old, new)
}

func AtomicLoadU64(ptr *uint64) uint64 {
	return atomic.LoadUint64(ptr)
}

func AtomicStoreU64(ptr *uint64, val uint64) {
	atomic.StoreUint64(ptr, val)
}

func AtomicAddU64(ptr *uint64, delta uint64) uint64 {
	return atomic.AddUint64(ptr, delta)
}

func AtomicCASU64(ptr *uint64, old, new uint64) bool {
	return atomic.CompareAndSwapUint64(ptr, old, new)
}

func AtomicLoadPtr(ptr *unsafe.Pointer) unsafe.Pointer {
	return atomic.LoadPointer(ptr)
}

func AtomicStorePtr(ptr *unsafe.Pointer, val unsafe.Pointer) {
	atomic.StorePointer(ptr, val)
}

func AtomicCASPtr(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(ptr, old, new)
}

func AtomicIncU32(ptr *uint32) uint32 {
	return atomic.AddUint32(ptr, 1)
}

func AtomicIncU64(ptr *uint64) uint64 {
	return atomic.AddUint64(ptr, 1)
}

func AtomicDecU32(ptr *uint32) uint32 {
	return atomic.AddUint32(ptr, ^uint32(0))
}

func AtomicDecU64(ptr *uint64) uint64 {
	return atomic.AddUint64(ptr, ^uint64(0))
}
