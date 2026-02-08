package gen

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const NullLevel = logrus.TraceLevel + 1

func int32a(u int32) string {
	return fmt.Sprintf("%c", u)
}

func parseDate2En(format string) string {
	f := []rune(format)
	var buf bytes.Buffer

	for _, r := range f {
		switch s := int32a(r); s {
		case "年", "Y":
			buf.WriteString("%Y")
		case "月", "m":
			buf.WriteString("%m")
		case "日", "d":
			buf.WriteString("%d")
		case "时", "H":
			buf.WriteString("%H")
		case "分", "M":
			buf.WriteString("%M")
		case "秒", "S":
			buf.WriteString("%S")
		default:
			buf.WriteString(s)
		}
	}
	return buf.String()
}

func parseFileFormat(serverName string, timeInterval int64) string {
	var buf bytes.Buffer
	buf.WriteString(serverName)
	buf.WriteString("_")

	cTime := time.Duration(timeInterval) * time.Second

	if cTime <= 0 || cTime >= time.Hour*24*365 {
		buf.WriteString("%Y%m%d.%H%M%S")
	} else if cTime >= time.Hour*24*30 {
		buf.WriteString("%Y%m")
	} else if cTime >= time.Hour*24 {
		buf.WriteString("%Y%m%d")
	} else if cTime >= time.Hour {
		buf.WriteString("%H_%M")
	} else if cTime >= time.Minute {
		buf.WriteString("%H_%M")
	} else if cTime >= time.Second {
		buf.WriteString("%H_%M_%S")
	}
	return buf.String()
}

func parseRootDir(serverName, root string) string {
	var buf bytes.Buffer
	buf.WriteString(root)
	buf.WriteString("/")
	if strings.Contains(serverName, "sport") || strings.Contains(serverName, "game") {
		buf.WriteString(serverName)
		buf.WriteString("/")
	}
	return buf.String()
}
