package main

import (
	"fmt"
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/anaminus/gxui/mixins"
	"github.com/anaminus/gxui/mixins/base"
	"github.com/anaminus/gxui/mixins/parts"
	"github.com/anaminus/gxui/themes/basic"
)

type faker struct {
	base.Control
	parts.Attachable

	ctxc            *ContextController
	window          *mixins.Window
	children        gxui.Children
	canvas          gxui.Canvas
	redrawRequested bool
}

func (f *faker) Init(theme gxui.Theme) {
	f.Control.Init(f, theme)
	f.OnAttach(func() {
		for _, v := range f.children {
			v.Control.Attach()
		}
	})
	f.OnDetach(func() {
		for _, v := range f.children {
			v.Control.Detach()
		}
	})
}

func (f *faker) DesiredSize(min, max math.Size) math.Size {
	return f.window.Size().Clamp(min, max)
}

func (f *faker) Paint(c gxui.Canvas) {
	s := f.window.Size().Contract(f.window.Padding()).Max(math.ZeroSize)
	for _, child := range f.children {
		child.Layout(child.Control.DesiredSize(math.ZeroSize, s).Rect())
	}
	for _, child := range f.children {
		if child.Control.IsVisible() {
			c.Push()
			c.AddClip(child.Control.Size().Rect().Offset(child.Offset))
			if cv := child.Control.Draw(); cv != nil {
				c.DrawCanvas(cv, child.Offset)
			}
			c.Pop()
		}
	}
	c.DrawRect(
		math.CreateRect(0, 0, c.Size().W, c.Size().H),
		gxui.Brush{Color: gxui.Color{0, 0, 0, 0.8}},
	)
}

func (f *faker) Redraw() {
	f.ctxc.Driver().AssertUIGoroutine()
	if !f.redrawRequested {
		f.redrawRequested = true
		f.ctxc.Window().Redraw()
	}
}

func (f *faker) Draw() gxui.Canvas {
	if !f.Attached() {
		panic(fmt.Errorf("Attempting to draw a non-attached control %T", f))
	}

	s := f.Size()
	if s.Area() == 0 {
		return nil // No area to draw in
	}
	if f.canvas == nil || f.canvas.Size() != s || f.redrawRequested {
		f.canvas = f.ctxc.Driver().CreateCanvas(s)
		f.redrawRequested = false
		f.Paint(f.canvas)
		f.canvas.Complete()
	}
	return f.canvas
}

func (f *faker) Children() gxui.Children {
	return f.children
}

func (f *faker) Finish() {
	f.ctxc.Window().RemoveAll()
	for _, child := range f.children {
		child.Control.SetParent(nil)
		f.ctxc.Window().AddChild(child.Control)
	}
}

func CreateFaker(ctxc *ContextController, visible bool) *faker {
	ctxc.Driver().AssertUIGoroutine()
	ctxc.Window().SetFocus(nil)
	f := &faker{
		ctxc:   ctxc,
		window: &ctxc.Window().(*basic.Window).Window,
	}
	f.Init(ctxc.Theme())

	f.children = make(gxui.Children, len(ctxc.Window().Children()))
	copy(f.children, ctxc.Window().Children())
	ctxc.Window().RemoveAll()
	for _, c := range f.children {
		c.Control.SetParent(f)
	}
	if visible {
		ctxc.Window().AddChild(f)
	}

	return f
}

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
	faker   *faker
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
	for _, control := range controls {
		if b, ok := control.(gxui.BubbleOverlay); ok {
			ctxi.bubbles = append(ctxi.bubbles, b)
		} else {
			ctxi.layout.AddChild(control)
		}
	}

	ctxi.faker = CreateFaker(c, ctx.IsDialog())

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
	ctxi.faker.Finish()
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
