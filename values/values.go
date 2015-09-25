package values

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/rbxplore/action"
	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
)

func GetWidget(ac *action.Controller, api *rbxdump.API, inst *rbxfile.Instance, prop string) gxui.Control {
	if ac == nil {
		return nil
	}
	switch inst.Properties[prop].(type) {
	case rbxfile.ValueString:
	case rbxfile.ValueBinaryString:
	case rbxfile.ValueProtectedString:
	case rbxfile.ValueContent:
	case rbxfile.ValueBool:
	case rbxfile.ValueInt:
	case rbxfile.ValueFloat:
	case rbxfile.ValueDouble:
	case rbxfile.ValueUDim:
	case rbxfile.ValueUDim2:
	case rbxfile.ValueRay:
	case rbxfile.ValueFaces:
	case rbxfile.ValueAxes:
	case rbxfile.ValueBrickColor:
	case rbxfile.ValueColor3:
	case rbxfile.ValueVector2:
	case rbxfile.ValueVector3:
	case rbxfile.ValueCFrame:
	case rbxfile.ValueToken:
	case rbxfile.ValueReference:
	case rbxfile.ValueVector3int16:
	case rbxfile.ValueVector2int16:
	case rbxfile.ValueNumberSequence:
	case rbxfile.ValueColorSequence:
	case rbxfile.ValueNumberRange:
	case rbxfile.ValueRect2D:
	}
	return nil
}
