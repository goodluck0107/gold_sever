package chanqueue

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"sync"
	"testing"
	"time"
)

type Server struct {
	wg    *sync.WaitGroup
	queue *Queue
}

func (s *Server) Close() {
	if s.queue != nil {
		fmt.Println(s.queue.Close())
	}
	s.wg.Wait()
	fmt.Println("wuhan close")
}

type QueueHandler struct {
	protoPool map[string]*Worker
}

func (Q *QueueHandler) OnInit() {
	Q.protoPool = make(map[string]*Worker)
	handler := func(v interface{}) error {
		msg := v.(*TestMsg)
		fmt.Printf("hello, my name is %s, age is %d .\n", msg.Name, msg.Age)
		if msg.Age < 5 {
			return fmt.Errorf("too young %d", msg.Age)
		} else {
			return nil
		}
	}
	Q.protoPool["proto_0"] = NewWorker(nil, handler)
	Q.protoPool["proto_1"] = NewWorker(&TestMsg{}, handler)
	Q.protoPool["proto_2"] = NewWorker(&TestMsg{}, nil)
	Q.protoPool["proto_3"] = NewWorker(&TestMsg{}, handler)
	Q.protoPool["proto_4"] = NewWorker(&TestMsg{}, nil)
	Q.protoPool["proto_5"] = NewWorker(&TestMsg{}, handler)
	Q.protoPool["proto_6"] = NewWorker(&TestMsg{}, nil)
	Q.protoPool["proto_7"] = NewWorker(&TestMsg{}, handler)
	Q.protoPool["proto_8"] = NewWorker(nil, handler)
}

func (Q *QueueHandler) WorkerByProto(proto string) (*Worker, bool) {
	worker, ok := Q.protoPool[proto]
	return worker, ok
}

type TestMsg struct {
	Name string
	Age  int
}

func TestQueue(t *testing.T) {
	server := new(Server)
	server.wg = new(sync.WaitGroup)
	server.queue = new(Queue)

	server.queue = NewQueue(&QueueHandler{}, 10)

	go server.queue.Run(server.wg)

	ticker := time.NewTicker(time.Second)
	var times int64
	for {
		<-ticker.C
		times++
		if times < 10 {
			rdm := static.HF_GetRandom(10)
			server.queue.Push(NewAsyncMsg(fmt.Sprintf("proto_%d", rdm), &TestMsg{
				Name: "hexu",
				Age:  rdm,
			}))
		} else {
			break
		}
	}
	ticker.Stop()
	server.Close()
}
