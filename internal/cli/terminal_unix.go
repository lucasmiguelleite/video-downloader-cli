//go:build unix

package cli

import (
	"os"

	"golang.org/x/sys/unix"
)

// IsRunningInBackground reports whether stdin's process group is backgrounded.
func IsRunningInBackground(stdin *os.File) (bool, error) {
	foregroundProcessGroupID, err := unix.IoctlGetInt(int(stdin.Fd()), unix.TIOCGPGRP)
	if err != nil {
		if err == unix.ENOTTY || err == unix.EINVAL {
			return false, nil
		}

		return false, err
	}

	return foregroundProcessGroupID != unix.Getpgrp(), nil
}
