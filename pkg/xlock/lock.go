package xlock

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"runtime/debug"
	"sync"
	"time"
)

var DEVAULT_TIMEOUT = time.Second * 3

type RWMutex struct {
	sync.RWMutex
	closechan chan struct{}
}

func (l *RWMutex) LockWithTimeOut(n time.Duration) bool {
	ticker := time.NewTicker(n)
	defer ticker.Stop()
	lockCh := make(chan struct{})
	go func() {
		l.Lock()
		lockCh <- struct{}{}
	}()
	select {
	case <-lockCh:
		return true
	case <-ticker.C:
		return false
	}

}

func (l *RWMutex) RlockWithTimeOut(n time.Duration) bool {
	ticker := time.NewTicker(n)
	defer ticker.Stop()
	lockCh := make(chan struct{})
	go func() {
		l.RLock()
		lockCh <- struct{}{}
	}()
	select {
	case <-lockCh:
		return true
	case <-ticker.C:
		return false
	}
}

func (l *RWMutex) RLockWithLog() {
	l.RLock()
	return
}

func (l *RWMutex) CustomLock() {

	l.Lock() //

	go func(stack string) {
	loop:
		for {
			select {
			case <-time.After(5 * time.Second):
				xlog.Logger().Errorf("get lock time out: %s", stack)
				break loop
			case <-l.closechan:
				break loop
			}
		}
	}(string(debug.Stack()))
	l.closechan = make(chan struct{})
	return
}

func (l *RWMutex) CustomUnLock() {
	close(l.closechan)
	l.Unlock()
	return
}
