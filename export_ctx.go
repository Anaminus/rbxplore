package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"os"
	"path/filepath"
)

type ExportContext struct {
	File      string
	Format    Format
	Minified  bool
	Finished  func(bool)
	ok        bool
	minButton gxui.Button
}

func (c *ExportContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()
	bubble := theme.CreateBubbleOverlay()

	dialog := CreateDialog(theme)
	dialog.SetTitle("Save As...")
	actionExport := dialog.AddAction("Save", true, func() {
		if c.File == "" || c.Format == FormatNone {
			return
		}
		if _, err := os.Stat(c.File); !os.IsNotExist(err) {
			ctxc.EnterContext(&AlertContext{
				Title:   "Confirm Overwrite",
				Text:    filepath.Base(c.File) + " already exists.\nWould you like to replace it?",
				Buttons: ButtonsYesNo,
				Finished: func(ok bool) {
					if ok {
						c.ok = true
						ctxc.ExitContext()
					}
				},
			})
		} else {
			c.ok = true
			ctxc.ExitContext()
		}
	})
	dialog.AddAction("Cancel", true, func() {
		c.ok = false
		ctxc.ExitContext()
	})
	setCanExport := func() {
		dialog.SetActionEnabled(actionExport, c.File != "" && c.Format != FormatNone)
	}
	setCanExport()

	layout := theme.CreateLinearLayout()
	layout.SetDirection(gxui.LeftToRight)
	dialog.Container().AddChild(layout)
	left := theme.CreateLinearLayout()
	left.SetPadding(math.Spacing{0, 0, 20, 0})
	layout.AddChild(left)
	right := theme.CreateLinearLayout()
	right.SetPadding(math.Spacing{20, 0, 0, 0})
	layout.AddChild(right)

	// File
	{
		label := theme.CreateLabel()
		label.SetText("File")
		label.SetMargin(math.Spacing{0, 8, 0, 8})
		left.AddChild(label)

		layout := theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.SetVerticalAlignment(gxui.AlignMiddle)

		textbox := theme.CreateTextBox()
		textbox.SetDesiredWidth(300)
		textbox.SetText(c.File)
		textbox.OnTextChanged(func([]gxui.TextBoxEdit) {
			c.File = textbox.Text()
			setCanExport()
		})
		layout.AddChild(textbox)

		button := CreateButton(theme, "Select...")
		button.OnClick(func(gxui.MouseEvent) {
			selectCtx := &FileSelectContext{
				SelectedFile: c.File,
				Saving:       true,
			}
			selectCtx.Finished = func() {
				if selectCtx.SelectedFile == "" {
					return
				}
				textbox.SetText(selectCtx.SelectedFile)
				c.File = selectCtx.SelectedFile
				setCanExport()
			}
			ctxc.EnterContext(selectCtx)
		})
		layout.AddChild(button)

		right.AddChild(layout)

	}

	// Format
	{
		label := theme.CreateLabel()
		label.SetText("Format")
		label.SetMargin(math.Spacing{0, 8, 0, 8})
		left.AddChild(label)

		layout := theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.SetVerticalAlignment(gxui.AlignMiddle)

		dropdown := theme.CreateDropDownList()
		dropdown.SetAdapter(new(FormatAdapter))
		dropdown.SetBubbleOverlay(bubble)
		dropdown.SetPadding(math.Spacing{5, 5, 5, 5})
		dropdown.SetMargin(math.Spacing{3, 3, 3, 3})
		dropdown.OnSelectionChanged(func(item gxui.AdapterItem) {
			if format, ok := item.(Format); !ok {
				dropdown.Select(c.Format)
				return
			} else {
				c.Format = format
				setCanExport()
			}
		})
		dropdown.Select(c.Format)
		layout.AddChild(dropdown)

		c.minButton = CreateButton(theme, "Minified")
		c.minButton.SetType(gxui.ToggleButton)
		c.minButton.SetChecked(c.Minified)
		layout.AddChild(c.minButton)

		right.AddChild(layout)
	}

	return []gxui.Control{dialog.Control(), bubble}, true
}

func (c *ExportContext) Exiting(ctxc *ContextController) {
	c.Minified = c.minButton.IsChecked()
	if c.Finished != nil {
		c.Finished(c.ok)
	}
}

func (c *ExportContext) IsDialog() bool {
	return true
}

func (c *ExportContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *ExportContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignCenter
}

func (c *ExportContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignMiddle
}
