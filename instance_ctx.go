package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
	"sort"
)

var templateInstance struct {
	className, name string
	service, props  bool
}

type classList struct {
	gxui.AdapterBase
	list []*rbxdump.Class
	size math.Size
}

func (l classList) Len() int {
	return len(l.list)
}
func (l classList) Less(i, j int) bool {
	return l.list[i].Name < l.list[j].Name
}
func (l classList) Swap(i, j int) {
	l.list[i], l.list[j] = l.list[j], l.list[i]
}

// Count returns the total number of items.
func (l *classList) Count() int {
	return len(l.list)
}

// ItemAt returns the AdapterItem for the item at index i. It is important
// for the Adapter to return consistent AdapterItems for the same data, so
// that selections can be persisted, or re-ordering animations can be played
// when the dataset changes.
// The AdapterItem returned must be equality-unique across all indices.
func (l *classList) ItemAt(index int) gxui.AdapterItem {
	return l.list[index]
}

// ItemIndex returns the index of item, or -1 if the adapter does not contain
// item.
func (l *classList) ItemIndex(item gxui.AdapterItem) int {
	class, _ := item.(*rbxdump.Class)
	if class == nil {
		return -1
	}
	for i, c := range l.list {
		if c == class {
			return i
		}
	}
	return -1
}

// Create returns a Control visualizing the item at the specified index.
func (l *classList) Create(theme gxui.Theme, index int) gxui.Control {
	class := l.list[index]
	label := theme.CreateLabel()
	label.SetText(class.Name)
	return label
}

// Size returns the size that each of the item's controls will be displayed
// at for the given theme.
func (l *classList) Size(gxui.Theme) math.Size {
	return l.size
}

func (l *classList) computeSize(theme gxui.Theme) {
	s := math.Size{}
	font := theme.DefaultFont()
	for _, class := range l.list {
		s = s.Max(font.Measure(&gxui.TextBlock{
			Runes: []rune(class.Name),
		}))
	}
	s.H = 22
	l.size = s
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
			shown := false
			button.OnClick(func(gxui.MouseEvent) {
				if shown {
					return
				}
				shown = true
				if Data.API == nil {
					return
				}
				classes := &classList{
					list: make([]*rbxdump.Class, len(Data.API.Classes)),
				}
				i := 0
				for _, class := range Data.API.Classes {
					classes.list[i] = class
					i++
				}
				sort.Sort(classes)
				classes.computeSize(theme)
				list := theme.CreateList()
				list.SetMargin(math.Spacing{0, 0, 0, 0})
				list.SetAdapter(classes)
				hide := func() {
					shown = false
					bubble.Hide()
					if c.className.Attached() {
						gxui.SetFocus(c.className)
					}
				}
				list.OnItemClicked(func(e gxui.MouseEvent, item gxui.AdapterItem) {
					hide()
					class, _ := item.(*rbxdump.Class)
					if class != nil {
						c.className.SetText(class.Name)
					}
				})
				list.OnKeyPress(func(ev gxui.KeyboardEvent) {
					switch ev.Key {
					case gxui.KeyEscape:
						hide()
					}
				})
				list.OnLostFocus(hide)
				button.OnDetach(hide)
				bubble.Show(list, gxui.TransformCoordinate(math.Point{X: 0, Y: button.Size().H}, button, bubble))
				gxui.SetFocus(list)
			})
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
