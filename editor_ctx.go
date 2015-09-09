package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/robloxapi/rbxfile"
)

type instanceNode struct {
	*rbxfile.Instance
	tooltips *gxui.ToolTipController
}

func (inst instanceNode) Count() int {
	return len(inst.GetChildren())
}

func (inst instanceNode) NodeAt(index int) gxui.TreeNode {
	return instanceNode{
		Instance: inst.GetChildren()[index],
		tooltips: inst.tooltips,
	}
}

func (inst instanceNode) ItemIndex(item gxui.AdapterItem) int {
	instItem := item.(*rbxfile.Instance)
loop:
	for {
		switch instItem.Parent() {
		case nil:
			return -1
		case inst.Instance:
			break loop
		}
		instItem = instItem.Parent()
	}
	for i, child := range inst.GetChildren() {
		if child == instItem {
			return i
		}
	}
	return -1
}

func (inst instanceNode) Item() gxui.AdapterItem {
	return inst.Instance
}

func (inst instanceNode) Create(theme gxui.Theme) gxui.Control {
	label := theme.CreateLabel()
	label.SetText(inst.Name())
	if inst.tooltips != nil {
		inst.tooltips.AddToolTip(label, 0.25, func(point math.Point) gxui.Control {
			tip := theme.CreateLabel()
			tip.SetText("Class: " + inst.ClassName)
			return tip
		})
	}

	if len(IconTextures) == 0 {
		return label
	}
	texture, ok := IconTextures[inst.ClassName]
	if !ok {
		texture = IconTextures[""]
	}
	icon := theme.CreateImage()
	icon.SetTexture(texture)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.SetVerticalAlignment(gxui.AlignMiddle)
	layout.AddChild(icon)
	layout.AddChild(label)
	return layout
}

type rootAdapter struct {
	gxui.AdapterBase
	*rbxfile.Root
	tooltips *gxui.ToolTipController
}

func (a rootAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 22}
}

func (root rootAdapter) Count() int {
	if root.Root == nil {
		return 0
	}
	return len(root.Instances)
}

func (root rootAdapter) NodeAt(index int) gxui.TreeNode {
	if root.Root == nil {
		return nil
	}
	return instanceNode{
		Instance: root.Instances[index],
		tooltips: root.tooltips,
	}
}

func (root rootAdapter) ItemIndex(item gxui.AdapterItem) int {
	if root.Root == nil {
		return -1
	}
	instItem := item.(*rbxfile.Instance)
	for instItem.Parent() != nil {
		instItem = instItem.Parent()
	}
	for i, inst := range root.Instances {
		if inst == instItem {
			return i
		}
	}
	return -1
}

func (root rootAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	if root.Root == nil {
		return nil
	}
	l := theme.CreateLabel()
	l.SetText(root.Instances[index].Name())
	return l
}

type propNode struct {
	inst *rbxfile.Instance
	name string
}

type propNodes []propNode

func (p propNodes) Len() int {
	return len(p)
}

func (p propNodes) Less(i, j int) bool {
	a, b := p[i], p[j]
	if a.inst == b.inst {
		return a.name < b.name
	} else {
		return a.inst.Name() < b.inst.Name()
	}
}

func (p propNodes) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type propsAdapter struct {
	gxui.AdapterBase
	props propNodes
}

func (p *propsAdapter) updateProps(inst *rbxfile.Instance) {
	if inst == nil {
		p.props = propNodes{}
		p.DataChanged(false)
		return
	}
	p.props = make(propNodes, 0, len(inst.Properties)+1)
	for name := range inst.Properties {
		p.props = append(p.props, propNode{
			inst: inst,
			name: name,
		})
	}
	sort.Sort(p.props)
	p.DataChanged(false)
}

func (p propsAdapter) Count() int {
	return len(p.props)
}

func (p propsAdapter) ItemAt(index int) gxui.AdapterItem {
	return p.props[index]
}

func (p propsAdapter) ItemIndex(item gxui.AdapterItem) int {
	for i, prop := range p.props {
		if prop == item.(propNode) {
			return i
		}
	}
	return -1
}

func (p propsAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	pr := p.props[index]
	l := theme.CreateLabel()
	prop := pr.inst.Properties[pr.name]
	v := prop.String()
	if len(v) < 128 {
		l.SetText(pr.name + " (" + prop.Type().String() + ") = " + prop.String())
	} else {
		l.SetText(pr.name + " (" + prop.Type().String() + ") = <long value>")
	}

	return l
}

func (p propsAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 22}
}

type EditorContext struct {
	session         *Session
	onChangeSession gxui.Event
	changeListener  gxui.EventSubscription
}

func (c *EditorContext) ChangeSession(s *Session) (err error) {
	if s != nil && s.File != "" {
		if err = s.DecodeFile(); err != nil {
			log.Printf("failed to decode session file: %s\n", err)
		}
	}
	if err == nil {
		c.session = s
	}
	c.onChangeSession.Fire(err)
	return
}

func (c *EditorContext) OnChangeSession(f func(error)) gxui.EventSubscription {
	if c.onChangeSession == nil {
		c.onChangeSession = gxui.CreateEvent(func(error) {})
	}
	return c.onChangeSession.Listen(f)
}

func (c *EditorContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()

	bubble := theme.CreateBubbleOverlay()
	tooltips := gxui.CreateToolTipController(bubble, ctxc.Driver())

	//// Menu
	menu := theme.CreateLinearLayout()
	menu.SetDirection(gxui.LeftToRight)

	actionButton := func(name string, f func(gxui.MouseEvent)) gxui.Button {
		button := CreateButton(theme, name)
		button.OnClick(f)
		menu.AddChild(button)
		return button
	}

	actionButton("New", func(e gxui.MouseEvent) {
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		if c.session == nil {
			c.ChangeSession(&Session{})
			return
		}
		if err := SpawnProcess("--new"); err != nil {
			log.Printf("failed to spawn process: %s\n", err)
		}
	})
	actionButton("Open", func(e gxui.MouseEvent) {
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		fmt.Println("TODO: enter Select context")
		session := &Session{
			File:   "",
			Format: FormatNone,
		}
		if c.session == nil {
			c.ChangeSession(session)
			return
		}
		fmt.Println("TODO: spawn process with selected file")
	})
	actionButton("Settings", func(e gxui.MouseEvent) {
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		fmt.Println("TODO: enter Settings context")
	})

	actionSave := actionButton("Save", func(e gxui.MouseEvent) {
		if c.session == nil {
			return
		}
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		fmt.Println("TODO: write session to file")
	})
	actionSaveAs := actionButton("Save As", func(e gxui.MouseEvent) {
		if c.session == nil {
			return
		}
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		fmt.Println("TODO: enter Export context")
		fmt.Println("TODO: change session output to selected file")
		fmt.Println("TODO: write session to selected file")
	})
	actionClose := actionButton("Close", func(e gxui.MouseEvent) {
		if c.session == nil {
			return
		}
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		c.ChangeSession(nil)
	})

	//// Editor
	rbxfiletree := theme.CreateTree()
	propsAdapter := &propsAdapter{}
	rbxfiletree.SetAdapter(&rootAdapter{
		tooltips: tooltips,
	})
	if c.changeListener != nil {
		c.changeListener.Unlisten()
	}
	c.changeListener = c.OnChangeSession(func(err error) {
		if err != nil {
			ctxc.EnterContext(&AlertContext{
				Title:   "Error",
				Text:    "Failed to open file: " + err.Error(),
				Buttons: ButtonsOK,
			})
			return
		}
		actionSave.SetVisible(c.session != nil)
		actionSaveAs.SetVisible(c.session != nil)
		actionClose.SetVisible(c.session != nil)

		propsAdapter.updateProps(nil)

		var root *rbxfile.Root
		if c.session != nil {
			root = c.session.Root
		}
		rbxfiletree.SetAdapter(&rootAdapter{
			Root:     root,
			tooltips: tooltips,
		})
	})

	propsList := theme.CreateList()
	propsList.SetAdapter(propsAdapter)

	rbxfiletree.OnSelectionChanged(func(item gxui.AdapterItem) {
		inst := item.(*rbxfile.Instance)
		propsAdapter.updateProps(inst)
	})

	rbxfiletree.ExpandAll()

	splitter := theme.CreateSplitterLayout()
	splitter.SetOrientation(gxui.Horizontal)
	splitter.AddChild(rbxfiletree)
	splitter.AddChild(propsList)

	//// Layout
	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.TopToBottom)
	layout.AddChild(menu)
	layout.AddChild(splitter)

	c.ChangeSession(c.session)

	return []gxui.Control{
		layout,
		bubble,
	}, true
}

func (c *EditorContext) Exiting(*ContextController) {
	if c.changeListener != nil {
		c.changeListener.Unlisten()
		c.changeListener = nil
	}
}

func (c *EditorContext) IsDialog() bool {
	return false
}
