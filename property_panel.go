package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/anaminus/rbxplore/action"
	"github.com/anaminus/rbxplore/values"
	"github.com/robloxapi/rbxfile"
	"sort"
)

type propItem struct {
	name   string
	label  gxui.Label
	widget gxui.Control
}

type PropertyPanel struct {
	Control    gxui.Control
	table      gxui.TableLayout
	theme      gxui.Theme
	ac         *action.Controller
	itemHeight int
	divider    float64
	instance   *rbxfile.Instance
}

func (p *PropertyPanel) relayout() {
	if p.ac != nil {
		p.ac.Lock()
		defer p.ac.Unlock()
	}
	p.table.RemoveAll()
	if p.instance == nil {
		p.table.SetGrid(2, 0)
		p.table.SetDesiredSize(math.Size{W: math.MaxSize.W, H: 0})
		p.redraw()
		return
	}

	p.table.SetGrid(2, len(p.instance.Properties))
	propNames := make([]string, 0, len(p.instance.Properties))
	for name := range p.instance.Properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	for i, name := range propNames {
		label := p.theme.CreateLabel()
		label.SetText(name)
		p.table.SetChildAt(0, i, 1, 1, label)

		widget := values.GetWidget(p.ac, Data.API, p.instance, name)
		if widget == nil {
			label := p.theme.CreateLabel()
			label.SetColor(gxui.Color{0.5, 0.5, 0.5, 1})
			if v := p.instance.Properties[name].String(); len(v) < 128 {
				label.SetText(v)
			} else {
				label.SetText("<long value>")
			}
			widget = label
		}
		p.table.SetChildAt(1, i, 1, 1, widget)
	}
	p.redraw()
}

func (p *PropertyPanel) redraw() {
	_, r := p.table.Grid()
	p.table.SetDesiredSize(math.Size{W: math.MaxSize.W, H: r * p.itemHeight})

	w := float64(p.table.Size().W)
	p.table.SetColumnWeight(0, int(w*p.divider))
	p.table.SetColumnWeight(1, int(w*(1-p.divider)))
}

func (p *PropertyPanel) SetInstance(inst *rbxfile.Instance) {
	if inst != p.instance {
		p.instance = inst
		p.relayout()
	}
}

func (p *PropertyPanel) SetActionController(ac *action.Controller) {
	if ac != p.ac {
		p.ac = ac
		p.relayout()
	}
}

func CreatePropertyPanel(theme gxui.Theme) *PropertyPanel {
	table := theme.CreateTableLayout()
	table.SetSizeClamped(true, false)
	scroll := theme.CreateScrollLayout()
	scroll.SetScrollAxis(false, true)
	scroll.SetChild(table)
	panel := &PropertyPanel{
		Control:    scroll,
		table:      table,
		theme:      theme,
		divider:    0.5,
		itemHeight: 26,
	}
	scroll.SetScrollLength(panel.itemHeight)
	panel.relayout()
	return panel
}
