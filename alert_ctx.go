package main

import (
	"github.com/anaminus/gxui"
	"strings"
)

type DialogButtons byte

const (
	ButtonsOK DialogButtons = iota
	ButtonsOKCancel
	ButtonsYesNo
	ButtonsYesNoCancel
)

type AlertContext struct {
	Title    string
	Text     string
	Buttons  DialogButtons
	Finished func(ok, cancel bool)
	dialog   Dialog
	ctxc     *ContextController
	ok       bool
	cancel   bool
}

func (c *AlertContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()

	c.dialog = CreateDialog(theme)
	c.dialog.SetTitle(c.Title)

	for _, s := range strings.Split(c.Text, "\n") {
		label := theme.CreateLabel()
		label.SetText(s)
		c.dialog.Container().AddChild(label)
	}

	switch c.Buttons {
	case ButtonsOK:
		c.dialog.AddAction("OK", true, func() {
			c.ok = true
			c.cancel = false
			ctxc.ExitContext()
		})
	case ButtonsOKCancel:
		c.dialog.AddAction("OK", true, func() {
			c.ok = true
			c.cancel = false
			ctxc.ExitContext()
		})
		c.dialog.AddAction("Cancel", true, func() {
			c.ok = false
			c.cancel = true
			ctxc.ExitContext()
		})
	case ButtonsYesNo:
		c.dialog.AddAction("Yes", true, func() {
			c.ok = true
			c.cancel = false
			ctxc.ExitContext()
		})
		c.dialog.AddAction("No", true, func() {
			c.ok = false
			c.cancel = false
			ctxc.ExitContext()
		})
	case ButtonsYesNoCancel:
		c.dialog.AddAction("Yes", true, func() {
			c.ok = true
			c.cancel = false
			ctxc.ExitContext()
		})
		c.dialog.AddAction("No", true, func() {
			c.ok = false
			c.cancel = false
			ctxc.ExitContext()
		})
		c.dialog.AddAction("Cancel", true, func() {
			c.ok = false
			c.cancel = true
			ctxc.ExitContext()
		})
	}
	return []gxui.Control{c.dialog.Control()}, true
}

func (c *AlertContext) Exiting(*ContextController) {
	if c.Finished != nil {
		c.Finished(c.ok, c.cancel)
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
