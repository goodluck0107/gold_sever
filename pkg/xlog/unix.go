// +build !windows

package xlog

import (
	"os"
	"syscall"
)

func redirect(filePath string) error {
	if !_PanicRedirectUnix {
		return nil
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
}
