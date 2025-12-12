package memory

import (
	"syscall"
	"unsafe"
)

func Mlock(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	ptr := unsafe.Pointer(&buf[0])
	return syscall.Mlock((*[1 << 30]byte)(ptr)[:len(buf)])
}

func Munlock(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	ptr := unsafe.Pointer(&buf[0])
	return syscall.Munlock((*[1 << 30]byte)(ptr)[:len(buf)])
}

func MlockAll() error {
	return syscall.Mlockall(syscall.MCL_CURRENT | syscall.MCL_FUTURE)
}

func MunlockAll() error {
	return syscall.Munlockall()
}

func TouchPages(buf []byte) {
	pageSize := 4096
	for i := 0; i < len(buf); i += pageSize {
		buf[i] = 0
	}
}
