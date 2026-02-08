package chanqueue

import (
	"errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

var (
	ErrProtoTypeConflict = errors.New("proto type conflict")
	ErrHandlerNotFind    = errors.New("handler not find")
	ErrNilHandler        = errors.New("handler is nil")
)

type Worker struct {
	worker    HandlerFunc
	protoType reflect.Type
}

func NewWorker(v interface{}, worker HandlerFunc) *Worker {
	return &Worker{
		worker:    worker,
		protoType: reflect.TypeOf(v),
	}
}

type HandlerFunc func(v interface{}) error

func (h *Worker) check(msg *Msg) error {
	if h.worker == nil {
		return ErrNilHandler
	}
	if reflect.TypeOf(msg.v) != h.protoType {
		return ErrProtoTypeConflict
	}
	return nil
}

func (h *Worker) Handle(mq *Queue, wg *sync.WaitGroup, msg *Msg) {
	err := h.check(msg)
	if err != nil {
		mq.outputf(logrus.ErrorLevel, "queue msg check error. head:%s, error:%v",
			msg.ProtoHead(), err)
		return
	}

	do := func(worker HandlerFunc, q *Queue, w *sync.WaitGroup, m *Msg) {
		if w != nil {
			w.Add(1)
			defer w.Done()
		}
		handleError := h.worker(msg.v)
		if handleError != nil {
			mq.outputf(logrus.ErrorLevel, "queue msg handler error. head:%s, error:%v",
				msg.ProtoHead(), handleError)
		} else {
			mq.outputf(logrus.InfoLevel, "queue msg handler 1 succeed: %s", msg.ProtoHead())
		}
	}

	if msg.IsAsync() {
		go do(h.worker, mq, wg, msg)
	} else {
		do(h.worker, mq, nil, msg)
	}
}
