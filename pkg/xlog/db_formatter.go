package xlog

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

type TextFormatter struct {
}

func (dbf *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(entry.Message)
	return buf.Bytes(), nil
}
