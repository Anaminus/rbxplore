package event

import "sync"

type Connection interface {
	// Disconnect disconnects a listener function from an event, such that the
	// event will no longer call the function.
	Disconnect()
}

type Event interface {
	// Connect adds a function to the event as a listener.
	Connect(fn func(...interface{})) Connection

	// Fire calls all listeners with the given arguments.
	Fire(v ...interface{})

	unlisten(id int)
}

type connection struct {
	event Event
	id    int
}

func (c *connection) Disconnect() {
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

type asyncEvent struct {
	nextid    int
	listeners []listener
	mutex     sync.Mutex
}

func (e *asyncEvent) unlisten(id int) {
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

func (e *asyncEvent) Connect(fn func(...interface{})) Connection {
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
	return &connection{e, id}
}

func (e *asyncEvent) Fire(v ...interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, l := range e.listeners {
		l.ch <- v
	}
}

type syncEvent struct {
	nextid    int
	listeners []listener
	mutex     sync.Mutex
}

func (e *syncEvent) unlisten(id int) {
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

func (e *syncEvent) Connect(fn func(...interface{})) Connection {
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
	}
	e.listeners = append(e.listeners, l)
	return &connection{e, id}
}

func (e *syncEvent) Fire(v ...interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for _, l := range e.listeners {
		l.fn(v...)
	}
}

// New returns a new Event. If sync is true, listeners are fired
// synchronously, in the order they were connected. Otherwise, they are fired
// asyncronously.
func New(sync bool) Event {
	if sync {
		return &syncEvent{}
	}
	return &asyncEvent{}
}
