package action

import (
	"errors"
	"sync"
)

type Action interface {
	// Setup sets up internal state necessary do and undo the action. This is
	// called once, before doing the action for the first time.
	Setup() error

	// Forward performs the action, either for the first time, or when redoing
	// the action.
	Forward() error

	// Backward performs the reverse of Forward, such that the state is as it
	// was before Forward was called.
	Backward() error
}

// An undo/redo stack with a circular buffer.
type historyStack struct {
	buffer          []Action
	head, tail, cur int
	ud, rd          bool
}

func (c *historyStack) Do(a Action) {
	if c.head == c.tail && c.ud {
		c.head++
		if c.head >= len(c.buffer) {
			c.head = 0
		}
	}
	if c.rd {
		if c.tail <= c.cur {
			for i := c.cur; i < len(c.buffer); i++ {
				c.buffer[i] = nil
			}
			for i := 0; i < c.tail; i++ {
				c.buffer[i] = nil
			}
		} else {
			for i := c.cur; i <= c.tail; i++ {
				c.buffer[i] = nil
			}
		}
		c.tail = c.cur
		c.rd = false
	}
	c.buffer[c.cur] = a
	c.ud = true
	c.cur++
	if c.cur >= len(c.buffer) {
		c.cur = 0
	}
	c.tail++
	if c.tail >= len(c.buffer) {
		c.tail = 0
	}
}

func (c *historyStack) Undo() (a Action) {
	if c.ud {
		c.cur--
		if c.cur < 0 {
			c.cur = len(c.buffer) - 1
		}
		a = c.buffer[c.cur]
		if c.cur == c.head {
			c.ud = false
		}
		c.rd = true
	}
	return
}

func (c *historyStack) Redo() (a Action) {
	if c.rd {
		a = c.buffer[c.cur]
		c.cur++
		if c.cur >= len(c.buffer) {
			c.cur = 0
		}
		if c.cur == c.tail {
			c.rd = false
		}
		c.ud = true
	}
	return
}

var NoAction = errors.New("no action")

type Controller struct {
	sync.Mutex
	stack *historyStack
}

func CreateController(historySize int) *Controller {
	return &Controller{
		stack: &historyStack{
			buffer: make([]Action, historySize),
		},
	}
}

func (ac *Controller) Do(a Action) error {
	ac.Lock()
	defer ac.Unlock()

	if a == nil {
		return NoAction
	}
	if err := a.Setup(); err != nil {
		return err
	}
	if err := a.Forward(); err != nil {
		return err
	}
	ac.stack.Do(a)
	return nil
}

func (ac *Controller) Undo() error {
	ac.Lock()
	defer ac.Unlock()

	a := ac.stack.Undo()
	if a == nil {
		return NoAction
	}
	return a.Backward()
}

func (ac *Controller) Redo() error {
	ac.Lock()
	defer ac.Unlock()

	a := ac.stack.Redo()
	if a == nil {
		return NoAction
	}
	return a.Forward()
}
