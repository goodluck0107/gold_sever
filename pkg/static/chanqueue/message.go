package chanqueue

import "reflect"

// Msg is a message basic
type Msg struct {
	// Head proto msg head, like a flag/tag
	head string
	// Async whether to have the message handler execute asynchronously
	async bool
	// V proto data, the data required by a message handler
	v interface{}
}

func NewMsg(head string, data interface{}) *Msg {
	return &Msg{
		head: head,
		v:    data,
	}
}

func NewAsyncMsg(head string, data interface{}) *Msg {
	return &Msg{
		head:  head,
		async: true,
		v:     data,
	}
}

func (m *Msg) ProtoHead() string {
	return m.head
}

func (m *Msg) ProtoData() interface{} {
	return m.v
}

func (m *Msg) ProtoType() reflect.Type {
	return reflect.TypeOf(m.v)
}

func (m *Msg) IsAsync() bool {
	return m.async
}
