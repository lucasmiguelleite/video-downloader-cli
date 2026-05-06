//go:build !unix

package cli

import "os"

func IsRunningInBackground(stdin *os.File) (bool, error) {
	return false, nil
}
