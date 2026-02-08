package object

import (
	"fmt"
	"sync"
	"time"
)

type Mu struct {
	tag string
	mu  sync.RWMutex
}

func (m *Mu) RLock() func() {
	m.mu.RLock()
	output(m.tag, "RLocked")
	start := time.Now()
	return func() {
		m.mu.RUnlock()
		output(m.tag, "RUnlocked ", expendTime(start))
	}
}

func (m *Mu) Lock() func() {
	m.mu.Lock()
	output(m.tag, "Locked")
	start := time.Now()
	return func() {
		m.mu.Unlock()
		output(m.tag, "Unlocked ", expendTime(start))
	}
}

func (m *Mu) setTag(t string) {
	m.tag = fmt.Sprintf("%2s ", t)
}

// expendTime return colorful string of expend time
func expendTime(start time.Time) string {
	return fmt.Sprintf("%.2fms", float64(time.Since(start).Nanoseconds()/1e4)/100.0)
}
