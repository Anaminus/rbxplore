package property

import (
	"github.com/anaminus/gxui"
	"github.com/robloxapi/rbxfile"
)

// Widget is used by a panel to modify a property of a selected instance. It
// is made to modify a value of a single type.
type widget interface {
	// Type returns the type of value that the widget modifies.
	Type() rbxfile.Type
	// Control returns a gxui.Control that modifies a property value.
	Control() gxui.Control
	// Value returns the value of the widget.
	Value() rbxfile.Value
	// SetValue sets the value of the widget, updating the display. This does
	// not count as the value being edited.
	SetValue(value rbxfile.Value)
	// OnEdited receives a function called after the value of the widget has
	// been changed by user interaction. Two values are passed: The
	// rbxfile.Value of the widget, and a bool indicating whether the edit is
	// a finalized action that should be commited to history. If the function
	// returns false, then the edit fails and the value is reverted to its
	// previous state.
	OnEdited(cb func(value rbxfile.Value, final bool) bool)
}

func createWidget(theme gxui.Theme, t rbxfile.Type) (w widget) {
	if theme == nil {
		return nil
	}
	switch t {
	case rbxfile.TypeString:
	case rbxfile.TypeBinaryString:
	case rbxfile.TypeProtectedString:
	case rbxfile.TypeContent:
	case rbxfile.TypeBool:
		w = &widgetBool{theme: theme}
	case rbxfile.TypeInt:
	case rbxfile.TypeFloat:
	case rbxfile.TypeDouble:
	case rbxfile.TypeUDim:
	case rbxfile.TypeUDim2:
	case rbxfile.TypeRay:
	case rbxfile.TypeFaces:
	case rbxfile.TypeAxes:
	case rbxfile.TypeBrickColor:
	case rbxfile.TypeColor3:
	case rbxfile.TypeVector2:
	case rbxfile.TypeVector3:
	case rbxfile.TypeCFrame:
	case rbxfile.TypeToken:
	case rbxfile.TypeReference:
	case rbxfile.TypeVector3int16:
	case rbxfile.TypeVector2int16:
	case rbxfile.TypeNumberSequence:
	case rbxfile.TypeColorSequence:
	case rbxfile.TypeNumberRange:
	case rbxfile.TypeRect2D:
	}
	return
}
