package event

import "sync"

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
	ch chan []interface{}
}

func (l listener) run() {
	for {
		if v, ok := <-l.ch; ok {
			l.fn(v...)
		} else {
			return
		}
	}
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
			close(l.ch)
			return
		}
	}
}

func (e *Event) Connect(fn func(...interface{})) *Connection {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if fn == nil {
		panic("received nil listener")
	}
	id := e.nextid
	e.nextid++
	l := listener{
		id: id,
		fn: fn,
		ch: make(chan []interface{}, 4),
	}
	go l.run()
	e.listeners = append(e.listeners, l)
	return &Connection{e, id}
}

func (e *Event) Fire(v ...interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, l := range e.listeners {
		l.ch <- v
	}
}
