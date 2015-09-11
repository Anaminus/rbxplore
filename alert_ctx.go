package main

import (
	"github.com/anaminus/gxui"
)

type AlertContext struct {
	Buttons  DialogButtons
	Title    string
	Text     string
	OK       bool
	dialog   Dialog
	ctxc     *ContextController
	Finished func(bool)
}

func (c *AlertContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()

	c.dialog = CreateDialog(theme)
	c.dialog.SetTitle(c.Title)

	label := theme.CreateLabel()
	label.SetText(c.Text)
	c.dialog.Container().AddChild(label)

	c.dialog.AddAction("OK", true, func() {
		c.OK = true
		ctxc.ExitContext()
	})
	if c.Buttons == ButtonsOKCancel {
		c.dialog.AddAction("Cancel", true, func() {
			c.OK = false
			ctxc.ExitContext()
		})
	}
	return []gxui.Control{c.dialog.Control()}, true
}

func (c *AlertContext) Exiting(*ContextController) {
	if c.Finished != nil {
		c.Finished(c.OK)
	}
}

func (c *AlertContext) IsDialog() bool {
	return true
}

func (c *AlertContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *AlertContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignCenter
}

func (c *AlertContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignMiddle
}
