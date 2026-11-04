//go:build !unix

package cli

import "os"

// IsRunningInBackground reports false on platforms without Unix process groups.
func IsRunningInBackground(stdin *os.File) (bool, error) {
	return false, nil
}
