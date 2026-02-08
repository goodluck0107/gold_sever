package chanqueue

import (
	"errors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
	"sync"
)

// Queue is a channel message queue
type Queue struct {
	handler  Handler
	messages chan *Msg

	disableLog bool
}

func NewQueue(handler Handler, size ...int) *Queue {
	q := &Queue{
		handler:  handler,
		messages: nil,
	}
	q.initChannel(size...)
	return q
}

func (mq *Queue) LogModel(output bool) {
	mq.disableLog = !output
}

// init initializes the message queue
func (mq *Queue) initChannel(size ...int) {
	if mq.messages == nil {
		if len(size) > 0 {
			mq.messages = make(chan *Msg, size[0])
		} else {
			const DefaultChannelBufferSize = 5000
			mq.messages = make(chan *Msg, DefaultChannelBufferSize)
		}
	}
}

func (mq *Queue) beforeRun() {
	if mq.messages == nil {
		mq.initChannel()
	}
	if mq.handler == nil {
		mq.handler = &defaultHandler{}
	}
	mq.handler.OnInit()
}

// close the message queue
func (mq *Queue) close() {
	close(mq.messages)
	mq.messages = nil
}

// Run initializes the message queue and the message handler,
// and immediately starts to continuously monitor the message queue.
// After the message is heard, it is sent to the
// corresponding handler for processing according to the message header,
// until the message queue is closed
func (mq *Queue) Run(wg *sync.WaitGroup) {
	defer func() {
		if x := recover(); x != nil {
			mq.outputf(logrus.ErrorLevel, "message queue panic: %+v", x)
		}
	}()

	mq.beforeRun()

	if wg != nil {
		wg.Add(1)
		defer wg.Done()
	}

	mq.output(logrus.WarnLevel, "message queue running...")

	for {
		msg := mq.pop()
		if msg == nil {
			continue
		}
		if msg.ProtoHead() == "" {
			break
		}
		mq.outputf(logrus.InfoLevel, "message queue pop a msg:%s, async:%t, data:%+v .", msg.ProtoHead(), msg.IsAsync(), msg.ProtoData())
		if queueHandler, ok := mq.handler.WorkerByProto(msg.ProtoHead()); ok {
			if queueHandler == nil {
				mq.outputf(logrus.ErrorLevel, "message queue has registered head:%s but handler is nil.",
					msg.ProtoHead())
			} else {
				queueHandler.Handle(mq, wg, msg)
			}
		} else {
			mq.outputf(logrus.ErrorLevel, "not to accept the queue msg. head:%s .",
				msg.ProtoHead())
		}
	}
	mq.close()
	mq.output(logrus.WarnLevel, "message queue stopped")
}

// Pop the first message in the message queue
func (mq *Queue) pop() *Msg {
	return <-mq.messages
}

// Push a message to the end of the message queue
func (mq *Queue) Push(msg *Msg) {
	if mq.messages == nil {
		mq.outputf(logrus.ErrorLevel, "push must after run.")
		return
	}
	mq.outputf(logrus.InfoLevel, "message queue push a msg:%s", msg.ProtoHead())
	mq.messages <- msg
}

// Close provides a method to close the message queue in order for external calls
func (mq *Queue) Close() error {
	if mq.messages == nil {
		return errors.New("queue not running")
	}
	mq.output(logrus.WarnLevel, "message queue closing...")
	mq.messages <- NewMsg("", nil)
	return nil
}

func (mq *Queue) output(lv logrus.Level, args ...interface{}) {
	if lv <= logrus.ErrorLevel {
		xlog.Logger().Error(args...)
	} else {
		if !mq.disableLog {
			switch lv {
			// case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			// 	syslog.Logger().Error(args...)
			case logrus.WarnLevel:
				xlog.Logger().Warning(args...)
			default:
				xlog.Logger().Info(args...)
			}
		}
	}
}

func (mq *Queue) outputf(lv logrus.Level, format string, args ...interface{}) {
	if lv <= logrus.ErrorLevel {
		xlog.Logger().Errorf(format, args...)
	} else {
		if !mq.disableLog {
			switch lv {
			// case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			// 	syslog.Logger().Errorf(format, args...)
			case logrus.WarnLevel:
				xlog.Logger().Warningf(format, args...)
			default:
				xlog.Logger().Infof(format, args...)
			}
		}
	}

}
