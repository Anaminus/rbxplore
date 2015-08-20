package main

import (
	"github.com/google/gxui"
)

type emptySub struct{}

func (emptySub) Unlisten() {}

// Absorber is a container that absorbs input.
type Absorber struct {
	gxui.Control
}

func CreateAbsorber(theme gxui.Theme, c gxui.Color, children ...gxui.Control) *Absorber {
	v := theme.CreateLinearLayout()
	v.SetBackgroundBrush(gxui.Brush{Color: c})
	v.SetSizeMode(gxui.Fill)
	v.SetDirection(gxui.TopToBottom)
	v.SetHorizontalAlignment(gxui.AlignCenter)

	for _, child := range children {
		v.AddChild(child)
	}

	return &Absorber{
		Control: v,
	}
}

func (a Absorber) Color(c gxui.Color) {
	a.Control.(gxui.LinearLayout).SetBackgroundBrush(gxui.Brush{Color: c})
}

func (a Absorber) IsMouseOver() bool                        { return false }
func (a Absorber) IsMouseDown(button gxui.MouseButton) bool { return false }

func (a Absorber) Click(gxui.MouseEvent) (consume bool)         { return true }
func (a Absorber) DoubleClick(gxui.MouseEvent) (consume bool)   { return true }
func (a Absorber) KeyPress(gxui.KeyboardEvent) (consume bool)   { return true }
func (a Absorber) KeyStroke(gxui.KeyStrokeEvent) (consume bool) { return true }
func (a Absorber) MouseScroll(gxui.MouseEvent) (consume bool)   { return true }

func (a Absorber) MouseMove(gxui.MouseEvent)    {}
func (a Absorber) MouseEnter(gxui.MouseEvent)   {}
func (a Absorber) MouseExit(gxui.MouseEvent)    {}
func (a Absorber) MouseDown(gxui.MouseEvent)    {}
func (a Absorber) MouseUp(gxui.MouseEvent)      {}
func (a Absorber) KeyDown(gxui.KeyboardEvent)   {}
func (a Absorber) KeyUp(gxui.KeyboardEvent)     {}
func (a Absorber) KeyRepeat(gxui.KeyboardEvent) {}

func (a Absorber) OnKeyPress(f func(gxui.KeyboardEvent)) gxui.EventSubscription   { return emptySub{} }
func (a Absorber) OnKeyStroke(f func(gxui.KeyStrokeEvent)) gxui.EventSubscription { return emptySub{} }
func (a Absorber) OnClick(f func(gxui.MouseEvent)) gxui.EventSubscription         { return emptySub{} }
func (a Absorber) OnDoubleClick(f func(gxui.MouseEvent)) gxui.EventSubscription   { return emptySub{} }
func (a Absorber) OnMouseMove(f func(gxui.MouseEvent)) gxui.EventSubscription     { return emptySub{} }
func (a Absorber) OnMouseEnter(f func(gxui.MouseEvent)) gxui.EventSubscription    { return emptySub{} }
func (a Absorber) OnMouseExit(f func(gxui.MouseEvent)) gxui.EventSubscription     { return emptySub{} }
func (a Absorber) OnMouseDown(f func(gxui.MouseEvent)) gxui.EventSubscription     { return emptySub{} }
func (a Absorber) OnMouseUp(f func(gxui.MouseEvent)) gxui.EventSubscription       { return emptySub{} }
func (a Absorber) OnMouseScroll(f func(gxui.MouseEvent)) gxui.EventSubscription   { return emptySub{} }
func (a Absorber) OnKeyDown(f func(gxui.KeyboardEvent)) gxui.EventSubscription    { return emptySub{} }
func (a Absorber) OnKeyUp(f func(gxui.KeyboardEvent)) gxui.EventSubscription      { return emptySub{} }
func (a Absorber) OnKeyRepeat(f func(gxui.KeyboardEvent)) gxui.EventSubscription  { return emptySub{} }
