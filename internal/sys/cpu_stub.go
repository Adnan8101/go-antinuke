//go:build !linux

package sys

func PinToCore(coreID int) error {
	return nil
}
