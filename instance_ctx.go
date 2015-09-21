package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
)

var templateInstance struct {
	className, name string
	service, props  bool
}

type InstanceContext struct {
	Instance        *rbxfile.Instance
	Finished        func(*rbxfile.Instance, bool)
	ok              bool
	className, name gxui.TextBox
	service, props  gxui.Button
}

func (c *InstanceContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()
	bubble := theme.CreateBubbleOverlay()

	dialog := CreateDialog(theme)
	dialog.SetTitle("Create Instance...")
	dialog.AddAction("Create", true, func() {
		c.ok = true
		ctxc.ExitContext()
	})
	dialog.AddAction("Cancel", true, func() {
		c.ok = false
		ctxc.ExitContext()
	})

	wrap := func(children ...gxui.Control) gxui.LinearLayout {
		layout := theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.SetHorizontalAlignment(gxui.AlignLeft)
		layout.SetVerticalAlignment(gxui.AlignMiddle)
		for _, child := range children {
			layout.AddChild(child)
		}
		return layout
	}

	table := theme.CreateTableLayout()
	table.SetGrid(2, 3)
	table.SetDesiredSize(math.Size{400, 28 * 3})
	table.SetSizeClamped(true, true)
	table.SetColumnWeight(1, 3)
	dialog.Container().AddChild(table)

	{
		label := theme.CreateLabel()
		label.SetText("Class Name:")
		table.SetChildAt(0, 0, 1, 1, wrap(label))
		c.className = theme.CreateTextBox()
		c.className.SetDesiredWidth(400)
		c.className.SetText(templateInstance.className)
		var l gxui.LinearLayout
		if Data.API != nil {
			button := theme.CreateButton()
			button.SetText("...")
			button.SetMargin(math.Spacing{0, 3, 3, 3})
			l = wrap(c.className, button)
		} else {
			l = wrap(c.className)
		}
		table.SetChildAt(1, 0, 1, 1, l)
	}
	{
		label := theme.CreateLabel()
		label.SetText("Name:")
		table.SetChildAt(0, 1, 1, 1, wrap(label))
		c.name = theme.CreateTextBox()
		c.name.SetDesiredWidth(400)
		c.name.SetText(templateInstance.name)
		table.SetChildAt(1, 1, 1, 1, wrap(c.name))
	}
	{
		c.service = CreateButton(theme, "Is service")
		c.service.SetType(gxui.ToggleButton)
		c.service.SetChecked(templateInstance.service)
		if Data.API != nil {
			c.props = CreateButton(theme, "Get properties from API")
			c.props.SetType(gxui.ToggleButton)
			c.props.SetChecked(templateInstance.props)
			table.SetChildAt(0, 2, 2, 1, wrap(c.service, c.props))
		} else {
			table.SetChildAt(0, 2, 2, 1, wrap(c.service))
		}
	}

	return []gxui.Control{dialog.Control(), bubble}, true
}

func (c *InstanceContext) Exiting(ctxc *ContextController) {
	if c.ok {
		templateInstance.className = c.className.Text()
		templateInstance.name = c.name.Text()
		templateInstance.service = c.service.IsChecked()
		if c.props != nil {
			templateInstance.props = c.props.IsChecked()
		}

		c.Instance = rbxfile.NewInstance(templateInstance.className, nil)
		if Data.API != nil && templateInstance.props {
			class := Data.API.Classes[c.Instance.ClassName]
			for class != nil {
				for _, member := range class.Members {
					if prop, ok := member.(*rbxdump.Property); ok {
						if prop.ReadOnly || prop.Deprecated || prop.Hidden {
							continue
						}
						switch prop.Name {
						case "Parent", "ClassName", "Archivable":
							continue
						}
						if enum := Data.API.Enums[prop.ValueType]; enum != nil {
							c.Instance.Set(prop.Name, rbxfile.NewValue(rbxfile.TypeToken))
						} else {
							c.Instance.Set(prop.Name, rbxfile.NewValue(rbxfile.TypeFromString(prop.ValueType)))
						}
					}
				}
				class = Data.API.Classes[class.Superclass]
			}
		}
		c.Instance.SetName(templateInstance.name)
		c.Instance.IsService = templateInstance.service
	}
	if c.Finished != nil {
		c.Finished(c.Instance, c.ok)
	}
}

func (c *InstanceContext) IsDialog() bool {
	return true
}

func (c *InstanceContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *InstanceContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignCenter
}

func (c *InstanceContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignMiddle
}
