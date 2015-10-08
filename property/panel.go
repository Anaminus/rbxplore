package property

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/anaminus/rbxplore/action"
	"github.com/anaminus/rbxplore/cmd"
	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
	"sort"
)

type Panel interface {
	Control() gxui.Control
	SetActionController(ac *action.Controller)
	SetAPI(api *rbxdump.API)
	SetInstance(inst *rbxfile.Instance)
	SetProperty(prop string, value rbxfile.Value)
}

type panel struct {
	control    gxui.Control
	table      gxui.TableLayout
	theme      gxui.Theme
	ac         *action.Controller
	api        *rbxdump.API
	itemHeight int
	divider    float64
	instance   *rbxfile.Instance
	widgets    []widget
}

func (p *panel) relayout() {
	if p.ac != nil {
		p.ac.Lock()
		defer p.ac.Unlock()
	}
	p.table.RemoveAll()
	for _, widget := range p.widgets {
		if widget != nil {
			widget.OnEdited(nil)
		}
	}
	if p.instance == nil {
		p.table.SetGrid(2, 0)
		p.table.SetDesiredSize(math.Size{W: math.MaxSize.W, H: 0})
		p.redraw()
		return
	}

	p.table.SetGrid(2, len(p.instance.Properties))
	p.widgets = make([]widget, len(p.instance.Properties))
	propNames := make([]string, 0, len(p.instance.Properties))
	for name := range p.instance.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	for i, name := range propNames {
		value := p.instance.Properties[name]
		label := p.theme.CreateLabel()
		label.SetText(name)
		p.table.SetChildAt(0, i, 1, 1, label)

		layout := p.theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.SetVerticalAlignment(gxui.AlignMiddle)
		layout.SetHorizontalAlignment(gxui.AlignLeft)

		widget := createWidget(p.theme, value.Type())
		if widget != nil {
			widget.SetValue(value)
			propName := name
			widget.OnEdited(func(value rbxfile.Value, final bool) bool {
				if !final {
					return true
				}
				p.ac.Do(cmd.SetProperty(p.instance, propName, value))
				// TODO: handle error
				return true
			})
			layout.AddChild(widget.Control())
		} else {
			label := p.theme.CreateLabel()
			label.SetColor(gxui.Color{0.5, 0.5, 0.5, 1})
			if v := value.String(); len(v) < 128 {
				label.SetText(v)
			} else {
				label.SetText("<long value>")
			}
			layout.AddChild(label)
		}
		p.widgets[i] = widget
		p.table.SetChildAt(1, i, 1, 1, layout)
	}
	p.redraw()
}

func (p *panel) redraw() {
	_, r := p.table.Grid()
	p.table.SetDesiredSize(math.Size{W: math.MaxSize.W, H: r * p.itemHeight})

	w := float64(p.table.Size().W)
	p.table.SetColumnWeight(0, int(w*p.divider))
	p.table.SetColumnWeight(1, int(w*(1-p.divider)))
}

func (p *panel) Control() gxui.Control {
	return p.control
}

func (p *panel) SetInstance(inst *rbxfile.Instance) {
	if inst != p.instance {
		p.instance = inst
		p.relayout()
	}
}

func (p *panel) SetActionController(ac *action.Controller) {
	if ac != p.ac {
		p.ac = ac
		p.relayout()
	}
}

func (p *panel) SetAPI(api *rbxdump.API) {
	if api != p.api {
		p.api = api
		p.relayout()
	}
}

func (p *panel) SetProperty(prop string, value rbxfile.Value) {
}

func CreatePanel(theme gxui.Theme) Panel {
	table := theme.CreateTableLayout()
	table.SetSizeClamped(true, false)
	scroll := theme.CreateScrollLayout()
	scroll.SetScrollAxis(false, true)
	scroll.SetChild(table)
	panel := &panel{
		control:    scroll,
		table:      table,
		theme:      theme,
		divider:    0.5,
		itemHeight: 26,
	}
	scroll.SetScrollLength(panel.itemHeight)
	panel.relayout()
	return panel
}
