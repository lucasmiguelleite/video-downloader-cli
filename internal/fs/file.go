// Package fs provides functions for file operations, such as saving data to a file.
package fs

import "os"

func Save(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}
