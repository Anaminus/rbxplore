package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
)

type Dialog interface {
	Container() gxui.Container
	Control() gxui.Control
	Title() (title string)
	SetTitle(title string)
	AddAction(text string, enabled bool, action func()) (id int)
	RemoveAction(id int)
	RemoveAllActions()
	SetActionEnabled(id int, enabled bool)
}

type dialogAction struct {
	button  gxui.Button
	label   gxui.Label
	enabled bool
	fn      func()
}

type dialog struct {
	control         gxui.Control
	container       gxui.Container
	title           gxui.Label
	nextid          int
	newButton       func() gxui.Button
	newLabel        func() gxui.Label
	actions         map[int]*dialogAction
	actionContainer gxui.Container
}

func (c *dialog) Title() string {
	return c.title.Text()
}

func (c *dialog) SetTitle(title string) {
	c.title.SetText(title)
}

func (c *dialog) Container() gxui.Container {
	return c.container
}

func (c *dialog) Control() gxui.Control {
	return c.control
}

func (c *dialog) AddAction(text string, enabled bool, action func()) int {
	id := c.nextid
	c.nextid++

	a := new(dialogAction)
	a.fn = action

	a.label = c.newLabel()
	a.label.SetText(text)

	a.button = c.newButton()
	a.button.AddChild(a.label)
	a.button.OnClick(func(gxui.MouseEvent) {
		if a.enabled {
			a.fn()
		}
	})

	c.actions[id] = a

	c.actionContainer.AddChild(a.button)
	c.setActionEnabled(id, enabled)
	return id
}

func (c *dialog) removeAction(id int) {
	c.actionContainer.RemoveChild(c.actions[id].button)
	delete(c.actions, id)
}

func (c *dialog) RemoveAction(id int) {
	c.removeAction(id)
}

func (c *dialog) RemoveAllActions() {
	for id := range c.actions {
		c.removeAction(id)
	}
}

func (c *dialog) setActionEnabled(id int, enabled bool) {
	action, ok := c.actions[id]
	if !ok {
		panic("unknown action button")
	}
	action.enabled = enabled
	label := action.label
	cl := label.Color()
	if enabled {
		label.SetColor(gxui.Color{cl.R, cl.G, cl.B, 1})
	} else {
		label.SetColor(gxui.Color{cl.R, cl.G, cl.B, 0.3})
	}
}

func (c *dialog) SetActionEnabled(id int, enabled bool) {
	c.setActionEnabled(id, enabled)
}

func CreateDialog(theme gxui.Theme) Dialog {
	c := new(dialog)
	c.actions = make(map[int]*dialogAction, 4)
	c.newButton = theme.CreateButton
	c.newLabel = theme.CreateLabel

	c.title = theme.CreateLabel()

	container := theme.CreateLinearLayout()
	container.SetDirection(gxui.TopToBottom)
	container.SetHorizontalAlignment(gxui.AlignLeft)
	container.SetVerticalAlignment(gxui.AlignTop)
	container.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
	c.container = container

	actionContainer := theme.CreateLinearLayout()
	actionContainer.SetDirection(gxui.LeftToRight)
	actionContainer.SetVerticalAlignment(gxui.AlignMiddle)
	actionContainer.SetHorizontalAlignment(gxui.AlignRight)
	c.actionContainer = actionContainer

	layout := theme.CreateLinearLayout()
	layout.SetBackgroundBrush(gxui.Brush{Color: gxui.Color{0.2, 0.2, 0.2, 1}})
	layout.SetDirection(gxui.TopToBottom)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.AddChild(c.title)
	layout.AddChild(container)
	layout.AddChild(actionContainer)
	c.control = layout

	return c
}
