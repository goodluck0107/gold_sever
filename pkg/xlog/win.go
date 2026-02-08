// +build windows

package xlog

import (
	"os"
	"syscall"
)

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

func redirect(filePath string) error {
	if !_PanicRedirectWin {
		return nil
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	err = _StderrRedirect(syscall.STD_ERROR_HANDLE, syscall.Handle(file.Fd()))
	if err != nil {
		return err
	}
	return nil
}

// stderr流重定向
func _StderrRedirect(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}
