package main

import (
	"github.com/anaminus/gxui"
)

type Context interface {
	Entering(*ContextController) ([]gxui.Control, bool)
	Exiting(*ContextController)
	IsDialog() bool
	Direction() gxui.Direction
	HorizontalAlignment() gxui.HorizontalAlignment
	VerticalAlignment() gxui.VerticalAlignment
}

type contextItem struct {
	context Context
	layout  gxui.LinearLayout
	bubbles []gxui.BubbleOverlay
}

type ContextController struct {
	driver gxui.Driver
	window gxui.Window
	theme  gxui.Theme
	stack  []contextItem
}

func (c *ContextController) createContextItem(ctx Context) (ctxi contextItem, ok bool) {
	controls, ok := ctx.Entering(c)
	if !ok {
		return ctxi, false
	}

	ctxi.context = ctx
	ctxi.layout = c.theme.CreateLinearLayout()
	ctxi.layout.SetSizeMode(gxui.Fill)
	ctxi.layout.SetDirection(ctx.Direction())
	ctxi.layout.SetHorizontalAlignment(ctx.HorizontalAlignment())
	ctxi.layout.SetVerticalAlignment(ctx.VerticalAlignment())
	if ctx.IsDialog() {
		ctxi.layout.SetBackgroundBrush(gxui.Brush{Color: gxui.Color{0, 0, 0, 0.8}})
	} else {
		ctxi.layout.SetBackgroundBrush(gxui.Brush{Color: gxui.Color{0, 0, 0, 1}})
	}
	for _, control := range controls {
		if b, ok := control.(gxui.BubbleOverlay); ok {
			ctxi.bubbles = append(ctxi.bubbles, b)
		} else {
			ctxi.layout.AddChild(control)
		}
	}

	return ctxi, true
}

func (c *ContextController) EnterContext(ctx Context) bool {
	if len(c.stack) < 1 {
		panic("empty context stack")
	}

	ctxi, ok := c.createContextItem(ctx)
	if !ok {
		return false
	}

	c.stack = append(c.stack, ctxi)
	c.window.AddChild(ctxi.layout)
	for _, bubble := range ctxi.bubbles {
		c.window.AddChild(bubble)
	}

	return true
}

func (c *ContextController) ExitContext() bool {
	if len(c.stack) < 1 {
		panic("empty context stack")
	}

	// Fail if there is only one context.
	if len(c.stack) <= 1 {
		return false
	}

	ctxi := c.stack[len(c.stack)-1]
	c.stack[len(c.stack)-1] = contextItem{}
	c.stack = c.stack[:len(c.stack)-1]

	c.window.RemoveChild(ctxi.layout)
	for _, bubble := range ctxi.bubbles {
		c.window.RemoveChild(bubble)
	}

	ctxi.context.Exiting(c)

	return true
}

func (c *ContextController) Driver() gxui.Driver {
	return c.driver
}

func (c *ContextController) Window() gxui.Window {
	return c.window
}

func (c *ContextController) Theme() gxui.Theme {
	return c.theme
}

func CreateContextController(driver gxui.Driver, window gxui.Window, theme gxui.Theme, ctx Context) (*ContextController, bool) {
	c := &ContextController{
		driver: driver,
		window: window,
		theme:  theme,
		stack:  make([]contextItem, 0, 4),
	}
	ctxi, ok := c.createContextItem(ctx)
	if !ok {
		return nil, false
	}
	c.stack = append(c.stack, ctxi)
	c.window.AddChild(ctxi.layout)
	for _, bubble := range ctxi.bubbles {
		c.window.AddChild(bubble)
	}
	return c, true
}
