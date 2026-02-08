package xlog

import (
	"testing"
	"time"
)

func TestElkLogger(t *testing.T) {
	ticker := time.NewTicker(time.Second * 3)
	for {
		<-ticker.C

		ELogger("GpsInfo").WithFields(map[string]interface{}{
			"Hour":   time.Now().Hour(),
			"Minute": time.Now().Minute(),
			"Second": time.Now().Second(),
		}).Error("GpsInfo")

		ELogger("UserInfo").WithFields(map[string]interface{}{
			"Hour":   time.Now().Hour(),
			"Minute": time.Now().Minute(),
			"Second": time.Now().Second(),
		}).Info("UserInfo")
	}
}
