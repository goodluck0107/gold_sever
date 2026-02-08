package xlog

import (
	"github.com/open-source/game/chess.git/pkg/xlog/rotate"
)

// 监听日志文件切割事件
type _EventHandler struct {
	inDocker bool
}

// 日志文件切割事件响应
// 当日志指向的文件改变时会调用该方法
// 目前该方法会紧随日志文件的变化重定向stderr流
func (e *_EventHandler) Handle(event rotate.Event) {
	if !e.inDocker {
		return
	}
	if event.Type() == rotate.FileRotatedEventType {
		if fevent, ok := event.(*rotate.FileRotatedEvent); ok {
			redirectStderr(fevent.CurrentFile())
		}
	}
}
