//go:build linux

package sys

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// PinToCore pins the current thread to the specified CPU core.
func PinToCore(coreID int) error {
	runtime.LockOSThread()

	var mask unix.CPUSet
	mask.Zero()
	mask.Set(coreID)

	return unix.SchedSetaffinity(0, &mask)
}
