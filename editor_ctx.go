package main

import (
	"fmt"
	"github.com/anaminus/rbxplore/action"
	"github.com/anaminus/rbxplore/event"
	"log"
	"path/filepath"
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
	return len(inst.Children)
}

func (inst instanceNode) NodeAt(index int) gxui.TreeNode {
	return instanceNode{
		Instance: inst.Children[index],
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
	for i, child := range inst.Children {
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

	if len(Data.Icons) == 0 {
		return label
	}
	texture, ok := Data.Icons[inst.ClassName]
	if !ok {
		texture = Data.Icons[""]
	}
	icon := theme.CreateImage()
	icon.SetMargin(math.Spacing{3, 0, 0, 0})
	icon.SetTexture(texture)

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.SetVerticalAlignment(gxui.AlignMiddle)
	layout.AddChild(icon)
	layout.AddChild(label)
	return layout
}

////////////////

type rootAdapter struct {
	gxui.AdapterBase
	*rbxfile.Root
	tooltips *gxui.ToolTipController
	ctx      *EditorContext
}

func (a rootAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 22}
}

func (root rootAdapter) Count() int {
	if root.Root == nil {
		return 0
	}
	return len(root.Instances) + 1
}

func (root rootAdapter) NodeAt(index int) gxui.TreeNode {
	if root.Root == nil {
		return nil
	}
	if index == len(root.Instances) {
		return addRootNode{
			root: root.Root,
			ctx:  root.ctx,
		}
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
	switch item := item.(type) {
	case addRootItem:
		if item.Root != root.Root {
			return -1
		}
		return len(root.Root.Instances)
	case *rbxfile.Instance:
		for item.Parent() != nil {
			item = item.Parent()
		}
		for i, inst := range root.Instances {
			if inst == item {
				return i
			}
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

////////////////

type addRootNode struct {
	root *rbxfile.Root
	ctx  *EditorContext
}

type addRootItem struct {
	*rbxfile.Root
}

func (node addRootNode) Count() int {
	return 0
}

func (node addRootNode) NodeAt(index int) gxui.TreeNode {
	return nil
}

func (node addRootNode) ItemIndex(item gxui.AdapterItem) int {
	return -1
}

func (node addRootNode) Item() gxui.AdapterItem {
	return addRootItem{Root: node.root}
}

func (node addRootNode) Create(theme gxui.Theme) gxui.Control {
	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	layout.SetHorizontalAlignment(gxui.AlignLeft)
	layout.SetVerticalAlignment(gxui.AlignMiddle)
	ctx := node.ctx
	{
		button := CreateButton(theme, "Add Instance")
		button.OnClick(func(gxui.MouseEvent) {
			ctx.ctxc.EnterContext(&InstanceContext{
				Finished: func(child *rbxfile.Instance, ok bool) {
					if !ok {
						return
					}
					if err := ctx.session.Action.Do(ctx.session.AddRootInstance(child)); err != nil {
						ctx.ctxc.EnterContext(&AlertContext{
							Title:   "Error",
							Text:    "Failed to add instance:\n" + err.Error(),
							Buttons: ButtonsOK,
						})
						return
					}
					if ctx.tree.Select(child) {
						ctx.tree.Show(child)
					}
				},
			})
		})
		layout.AddChild(button)
	}
	{
		button := CreateButton(theme, "Add Model")
		button.OnClick(func(gxui.MouseEvent) {
			loadModel(ctx.ctxc, func(children []*rbxfile.Instance) {
				ag := make(action.Group, len(children))
				var first *rbxfile.Instance
				for i, child := range children {
					if first == nil {
						first = child
					}
					ag[i] = ctx.session.AddRootInstance(child)
				}
				if err := ctx.session.Action.Do(ag); err != nil {
					ctx.ctxc.EnterContext(&AlertContext{
						Title:   "Error",
						Text:    "Failed to add objects:\n" + err.Error(),
						Buttons: ButtonsOK,
					})
					return
				}
				if first != nil && ctx.tree.Select(first) {
					ctx.tree.Show(first)
				}
			})
		})
		layout.AddChild(button)
	}
	return layout
}

func loadModel(ctxc *ContextController, f func([]*rbxfile.Instance)) {
	selectCtx := &FileSelectContext{
		Type: FileSelect,
	}
	selectCtx.Finished = func() {
		if selectCtx.SelectedFile == "" {
			return
		}
		s := &Session{
			File:   selectCtx.SelectedFile,
			Format: FormatNone,
		}
		if err := s.DecodeFile(); err != nil {
			ctxc.EnterContext(&AlertContext{
				Title:   "Error",
				Text:    "Failed to open file:\n" + err.Error(),
				Buttons: ButtonsOK,
			})
			return
		}
		if len(s.Root.Instances) == 0 {
			return
		}
		f(s.Root.Instances)
	}
	ctxc.EnterContext(selectCtx)

}

////////////////

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
	actionListener  event.Connection
	ctxc            *ContextController
	tree            gxui.Tree
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

func (c *EditorContext) updateWindowTitle(window gxui.Window) {
	if c.session == nil {
		window.SetTitle("rbxplore")
	} else if c.session.File == "" {
		window.SetTitle("(new file) - rbxplore")
	} else {
		window.SetTitle(filepath.Base(c.session.File) + " - rbxplore")
	}
}

func (c *EditorContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	c.ctxc = ctxc
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
			c.ChangeSession(NewSession())
			return
		}
		if Settings.Get("spawn_processes").(bool) {
			if err := SpawnProcess("--new"); err != nil {
				log.Printf("failed to spawn process: %s\n", err)
			}
			return
		}
		if c.session != nil {
			fmt.Println("TODO: prompt to save file")
		}
		c.ChangeSession(NewSession())
	})
	actionButton("Open", func(e gxui.MouseEvent) {
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		if c.session != nil {
			fmt.Println("TODO: prompt to save file")
		}
		selectCtx := &FileSelectContext{
			SelectedFile: "",
			Type:         FileOpen,
		}
		selectCtx.Finished = func() {
			if selectCtx.SelectedFile == "" {
				return
			}
			if c.session != nil && Settings.Get("spawn_processes").(bool) {
				if err := SpawnProcess(selectCtx.SelectedFile); err != nil {
					log.Printf("failed to spawn process: %s\n", err)
				}
				return
			}
			session := NewSession()
			session.File = selectCtx.SelectedFile
			session.Format = FormatNone
			c.ChangeSession(session)
		}
		if !ctxc.EnterContext(selectCtx) {
			return
		}
	})
	actionButton("Settings", func(e gxui.MouseEvent) {
		if e.Button != gxui.MouseButtonLeft {
			return
		}
		ctxc.EnterContext(&SettingsContext{})
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
		exportCtx := &ExportContext{
			File:     c.session.File,
			Format:   c.session.Format,
			Minified: c.session.Minified,
		}
		exportCtx.Finished = func(ok bool) {
			if ok {
				c.session.File = exportCtx.File
				c.session.Format = exportCtx.Format
				c.session.Minified = exportCtx.Minified
				c.updateWindowTitle(ctxc.Window())
				if err := c.session.EncodeFile(); err != nil {
					ctxc.EnterContext(&AlertContext{
						Title:   "Error",
						Text:    "Failed to save file: " + err.Error(),
						Buttons: ButtonsOK,
					})
				}
			}
		}
		ctxc.EnterContext(exportCtx)
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
	var updateSelection func(gxui.AdapterItem)
	c.tree = theme.CreateTree()
	c.tree.SetAdapter(&rootAdapter{
		tooltips: tooltips,
		ctx:      c,
	})
	if c.changeListener != nil {
		c.changeListener.Unlisten()
	}
	c.changeListener = c.OnChangeSession(func(err error) {
		if c.actionListener != nil {
			c.actionListener.Disconnect()
			c.actionListener = nil
		}
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

		c.updateWindowTitle(ctxc.Window())

		if updateSelection != nil {
			updateSelection(nil)
		}
		c.tree.Select(nil)

		var root *rbxfile.Root
		if c.session != nil {
			c.actionListener = c.session.Action.OnUpdate(func(...interface{}) {
				c.session.Action.Lock()
				ctxc.Driver().CallSync(func() {
					c.tree.Adapter().(*rootAdapter).DataChanged(false)
				})
				c.session.Action.Unlock()
			})
			root = c.session.Root
		}
		c.tree.SetAdapter(&rootAdapter{
			Root:     root,
			tooltips: tooltips,
			ctx:      c,
		})
	})

	propsLayout := theme.CreateLinearLayout()
	propsLayout.SetDirection(gxui.TopToBottom)
	propsLayout.SetHorizontalAlignment(gxui.AlignLeft)
	propsLayout.SetVerticalAlignment(gxui.AlignTop)

	propsButtons := theme.CreateLinearLayout()
	propsButtons.SetDirection(gxui.LeftToRight)
	propsButtons.SetHorizontalAlignment(gxui.AlignLeft)
	propsButtons.SetVerticalAlignment(gxui.AlignMiddle)
	propsLayout.AddChild(propsButtons)

	addChildButton := CreateButton(theme, "Add Child")
	addChildButton.SetVisible(false)
	addChildButton.OnClick(func(gxui.MouseEvent) {
		inst, _ := c.tree.Selected().(*rbxfile.Instance)
		if inst == nil {
			return
		}
		ctxc.EnterContext(&InstanceContext{
			Finished: func(child *rbxfile.Instance, ok bool) {
				if !ok {
					return
				}
				if err := c.session.Action.Do(c.session.SetParent(child, inst)); err != nil {
					ctxc.EnterContext(&AlertContext{
						Title:   "Error",
						Text:    "Failed to add instance:\n" + err.Error(),
						Buttons: ButtonsOK,
					})
					return
				}
				if c.tree.Select(child) {
					c.tree.Show(child)
				}
			},
		})
	})
	propsButtons.AddChild(addChildButton)

	addModelButton := CreateButton(theme, "Add Model")
	addModelButton.SetVisible(false)
	addModelButton.OnClick(func(gxui.MouseEvent) {
		inst, _ := c.tree.Selected().(*rbxfile.Instance)
		if inst == nil {
			return
		}
		loadModel(ctxc, func(children []*rbxfile.Instance) {
			ag := make(action.Group, len(children))
			for i, child := range children {
				ag[i] = c.session.SetParent(child, inst)
			}
			if err := c.session.Action.Do(ag); err != nil {
				ctxc.EnterContext(&AlertContext{
					Title:   "Error",
					Text:    "Failed to add objects:\n" + err.Error(),
					Buttons: ButtonsOK,
				})
				return
			}
		})
	})
	propsButtons.AddChild(addModelButton)

	propsAdapter := &propsAdapter{}
	propsList := theme.CreateList()
	propsList.SetAdapter(propsAdapter)
	propsLayout.AddChild(propsList)

	splitter := theme.CreateSplitterLayout()
	splitter.SetOrientation(gxui.Horizontal)
	splitter.AddChild(c.tree)
	splitter.AddChild(propsLayout)

	//// Layout
	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.TopToBottom)
	layout.AddChild(menu)
	layout.AddChild(splitter)

	updateSelection = func(item gxui.AdapterItem) {
		inst, _ := item.(*rbxfile.Instance)
		addChildButton.SetVisible(inst != nil)
		addModelButton.SetVisible(inst != nil)
		propsAdapter.updateProps(inst)
	}
	c.tree.OnSelectionChanged(updateSelection)

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

func (c *EditorContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *EditorContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignLeft
}

func (c *EditorContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignTop
}
