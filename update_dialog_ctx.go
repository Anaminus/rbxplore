package main

import (
	"fmt"
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"io"
)

type UpdateDialogContext struct {
	dialog         Dialog
	ctxc           *ContextController
	UpdateFinished func()
}

func (c *UpdateDialogContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	driver := ctxc.Driver()
	theme := ctxc.Theme()

	c.dialog = CreateDialog(theme)
	c.dialog.SetTitle("Updating Data...")

	label := theme.CreateLabel()
	pbar := theme.CreateProgressBar()
	pbar.SetDesiredSize(math.Size{300, 24})
	c.dialog.Container().AddChild(label)
	c.dialog.Container().AddChild(pbar)

	conn := Data.OnUpdateProgress(func(v ...interface{}) {
		driver.Call(func() {
			name := v[0].(string)
			progress := v[1].(int64)
			total := v[2].(int64)
			err, _ := v[3].(error)

			if err == io.EOF {
				label.SetText(fmt.Sprintf("Download of %s completed.", name))
			} else if err != nil {
			} else {
				if total < 0 {
					pbar.SetTarget(1)
					pbar.SetProgress(1)
					label.SetText(fmt.Sprintf("Downloading %s...", name))
				} else {
					pbar.SetTarget(int(total))
					pbar.SetProgress(int(progress))
					label.SetText(fmt.Sprintf("Downloading %s (%3.2f%%)...", name, float64(progress)/float64(total)*100))
				}
			}
		})
	})

	var canceled bool
	var actionCancel int
	actionCancel = c.dialog.AddAction("Cancel", true, func() {
		canceled = true
		conn.Disconnect()
		Data.CancelUpdate()
		label.SetText("Update canceled.")
		c.dialog.RemoveAction(actionCancel)
		c.dialog.AddAction("Close", true, func() {
			ctxc.ExitContext()
		})
	})

	go func() {
		err := Data.Update()
		driver.CallSync(func() {
			pbar.SetTarget(1)
			pbar.SetProgress(0)
			if canceled {
				return
			}
			conn.Disconnect()
			if err != nil {
				label.SetText(err.Error())
			} else {
				label.SetText("Update successful.")
			}
			c.dialog.RemoveAction(actionCancel)
			c.dialog.AddAction("Close", true, func() {
				ctxc.ExitContext()
			})
		})
	}()
	return []gxui.Control{c.dialog.Control()}, true
}

func (c *UpdateDialogContext) Exiting(*ContextController) {
	if c.UpdateFinished != nil {
		c.UpdateFinished()
	}
}

func (c *UpdateDialogContext) IsDialog() bool {
	return true
}
