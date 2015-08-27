package event

import (
	"sync"
)

type Connection struct {
	event *Event
	id    int
}

func (c *Connection) Disconnect() {
	if c.event != nil {
		c.event.unlisten(c.id)
		c.event = nil
	}
}

type listener struct {
	id int
	fn func(...interface{})
}

type Event struct {
	nextid    int
	listeners []listener
	mutex     sync.Mutex
}

func (e *Event) unlisten(id int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for i, l := range e.listeners {
		if l.id == id {
			copy(e.listeners[i:], e.listeners[i+1:])
			e.listeners = e.listeners[:len(e.listeners)-1]
			return
		}
	}
}

func (e *Event) Connect(l func(...interface{})) *Connection {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if l == nil {
		panic("received nil listener")
	}
	id := e.nextid
	e.nextid++
	e.listeners = append(e.listeners, listener{id, l})
	return &Connection{e, id}
}

func (e *Event) Fire(v ...interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, l := range e.listeners {
		go l.fn(v...)
	}
}
