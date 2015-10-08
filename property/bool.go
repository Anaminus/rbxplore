package property

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/robloxapi/rbxfile"
)

type widgetBool struct {
	theme   gxui.Theme
	control gxui.Control
	value   rbxfile.ValueBool

	image    gxui.Image
	onEdited func(value rbxfile.Value, final bool) bool
}

var checkMark = gxui.Polygon{
	gxui.PolygonVertex{Position: math.Point{X: 3, Y: 5}},
	gxui.PolygonVertex{Position: math.Point{X: 4, Y: 7}},
	gxui.PolygonVertex{Position: math.Point{X: 7, Y: 1}},
	gxui.PolygonVertex{Position: math.Point{X: 9, Y: 1}},
	gxui.PolygonVertex{Position: math.Point{X: 5, Y: 9}},
	gxui.PolygonVertex{Position: math.Point{X: 3, Y: 9}},
	gxui.PolygonVertex{Position: math.Point{X: 1, Y: 5}},
}

func (w *widgetBool) updateControl() {
	if w.control == nil {
		return
	}
	c := w.theme.Driver().CreateCanvas(math.Size{10, 10})
	if w.value {
		c.DrawPolygon(checkMark, gxui.TransparentPen, gxui.WhiteBrush)
	}
	c.Complete()
	w.image.SetCanvas(c)
}

func (w *widgetBool) Type() rbxfile.Type {
	return rbxfile.TypeBool
}

func (w *widgetBool) Control() gxui.Control {
	if w.control != nil {
		return w.control
	}
	button := w.theme.CreateButton()
	w.image = w.theme.CreateImage()
	button.AddChild(w.image)
	button.OnClick(func(gxui.MouseEvent) {
		if w.onEdited != nil && !w.onEdited(!w.value, true) {
			return
		}
		w.value = !w.value
		w.updateControl()
	})
	w.control = button
	w.updateControl()
	return w.control
}

func (w *widgetBool) Value() rbxfile.Value {
	return w.value
}

func (w *widgetBool) SetValue(value rbxfile.Value) {
	w.value = value.(rbxfile.ValueBool)
	w.updateControl()
}

func (w *widgetBool) OnEdited(cb func(value rbxfile.Value, final bool) bool) {
	w.onEdited = cb
}
