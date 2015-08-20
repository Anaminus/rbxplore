package main

import (
	"github.com/google/gxui"
)

type Context interface {
	Entering(gxui.Driver, gxui.Window, gxui.Theme) ([]gxui.Control, bool)
	Exiting()
	IsDialog() bool
	SetController(*ContextController)
}

type contextItem struct {
	context  Context
	controls []gxui.Control
	dialog   *Absorber
}

type ContextController struct {
	driver gxui.Driver
	window gxui.Window
	theme  gxui.Theme
	stack  []contextItem
}

func (c *ContextController) createContextItem(ctx Context) bool {
	controls, ok := ctx.Entering(c.driver, c.window, c.theme)
	if !ok {
		return false
	}
	ctxi := contextItem{
		context:  ctx,
		controls: controls,
	}
	if ctx.IsDialog() {
		ctxi.dialog = CreateAbsorber(c.theme, gxui.Color{0, 0, 0, 0.5}, ctxi.controls...)
	}
	ctx.SetController(c)
	c.stack = append(c.stack, ctxi)
	return true
}

func (c *ContextController) applyContext() {
	if len(c.stack) < 1 {
		panic("empty context stack")
	}
	ctxi := c.stack[len(c.stack)-1]
	if ctxi.dialog != nil {
		// Lay controls over current context, using a layout that blocks
		// input.
		c.window.AddChild(ctxi.dialog)
	} else {
		// Replace current controls with new controls.
		c.window.RemoveAll()
		for _, child := range ctxi.controls {
			c.window.AddChild(child)
		}
	}
}

func (c *ContextController) EnterContext(ctx Context) bool {
	if len(c.stack) < 1 {
		panic("empty context stack")
	}
	// Fail if current context is a dialog.
	if c.stack[len(c.stack)-1].dialog != nil {
		return false
	}
	if !c.createContextItem(ctx) {
		return false
	}
	c.applyContext()
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
	ctxi.context.SetController(nil)
	c.stack[len(c.stack)-1] = contextItem{}
	c.stack = c.stack[:len(c.stack)-1]

	if ctxi.dialog != nil {
		c.window.RemoveChild(ctxi.dialog)
	} else {
		c.applyContext()
	}

	return true
}

func CreateContextController(driver gxui.Driver, window gxui.Window, theme gxui.Theme, ctx Context) (*ContextController, bool) {
	c := &ContextController{
		driver: driver,
		window: window,
		theme:  theme,
		stack:  make([]contextItem, 0, 4),
	}
	if !c.createContextItem(ctx) {
		return c, false
	}
	c.applyContext()
	return c, true
}
