package object

import (
	"io"
	"log"
	"os"
)

type logger interface {
	Print(v ...interface{})
	SetOutput(w io.Writer)
}

var defaultLogger logger = log.New(os.Stdout, "", 0)

func SetLogger(l logger) {
	defaultLogger = l
}

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

func output(args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Print(args...)
	}
}
