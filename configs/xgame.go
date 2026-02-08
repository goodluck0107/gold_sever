package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	BIN_ENV  = "SRVBIN"
	BIN_PATH = "/home/game"
)

func main() {
	path, ok := syscall.Getenv(BIN_ENV)
	if !ok {
		fmt.Println(fmt.Sprintf("env %q not set, use `export %s=your bin path` to set it!", BIN_ENV, BIN_ENV))
		path = BIN_PATH
	}
	execPath := filepath.Join(path, "deploy.sh")
	fmt.Println("shell found: ", execPath)
	// fmt.Println(os.Args)
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("%s %s %s", execPath, OsArg(1), OsArg(2)))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// fmt.Println(cmd.String())
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Done")
}

func OsArg(i int) string {
	if i < len(os.Args) {
		return os.Args[i]
	}
	return ""
}
