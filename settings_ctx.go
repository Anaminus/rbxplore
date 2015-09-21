package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
)

type SettingsContext struct {
	settings map[string]interface{}
	ok       bool
	updated  bool
}

func (c *SettingsContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()

	group := func(name string, children ...gxui.Control) gxui.Control {
		container := theme.CreateLinearLayout()
		container.SetMargin(math.Spacing{L: 0, T: 0, B: 10, R: 0})
		label := theme.CreateLabel()
		label.SetText(name)
		container.AddChild(label)
		layout := theme.CreateLinearLayout()
		layout.SetDirection(gxui.LeftToRight)
		layout.SetPadding(math.Spacing{10, 10, 10, 10})
		layout.SetBorderPen(gxui.Pen{
			Width: 1,
			Color: gxui.Color{0.5, 0.5, 0.5, 1},
		})
		for _, child := range children {
			layout.AddChild(child)
		}
		container.AddChild(layout)
		return container
	}

	c.settings = Settings.Gets()

	items := []struct {
		name, file, url string
	}{
		{"ReflectionMetadata", "rmd_file", "rmd_update_url"},
		{"API Dump", "api_file", "api_update_url"},
		{"Icons", "icon_file", "icon_update_url"},
	}

	layout := theme.CreateLinearLayout()
	layout.SetHorizontalAlignment(gxui.AlignRight)

	// Files
	{
		table := theme.CreateTableLayout()
		table.SetGrid(2, len(items))
		table.SetDesiredSize(math.Size{-1, 32 * len(items)})
		table.SetColumnWeight(1, 3)
		for i, item := range items {
			file := item.file
			label := ctxc.Theme().CreateLabel()
			label.SetText(item.name)
			table.SetChildAt(0, i, 1, 1, label)

			layout := theme.CreateLinearLayout()
			layout.SetDirection(gxui.LeftToRight)
			layout.SetHorizontalAlignment(gxui.AlignLeft)
			layout.SetVerticalAlignment(gxui.AlignMiddle)

			textbox := theme.CreateTextBox()
			textbox.SetDesiredWidth(math.MaxSize.W)
			textbox.SetText(c.settings[file].(string))
			textbox.OnTextChanged(func([]gxui.TextBoxEdit) {
				c.settings[file] = textbox.Text()
			})
			layout.AddChild(textbox)

			button := CreateButton(theme, "Select...")
			button.OnClick(func(gxui.MouseEvent) {
				selectCtx := &FileSelectContext{
					SelectedFile: c.settings[file].(string),
				}
				selectCtx.Finished = func() {
					if selectCtx.SelectedFile == "" {
						return
					}
					textbox.SetText(selectCtx.SelectedFile)
					c.settings[item.file] = selectCtx.SelectedFile
				}
				ctxc.EnterContext(selectCtx)
			})
			layout.AddChild(button)

			table.SetChildAt(1, i, 1, 1, layout)
		}
		layout.AddChild(group("Files", table))
	}

	// Update URLs
	{
		table := theme.CreateTableLayout()
		table.SetGrid(2, len(items)+1)
		table.SetDesiredSize(math.Size{-1, 32 * (len(items) + 1)})
		table.SetColumnWeight(1, 3)
		for i, item := range items {
			url := item.url
			label := ctxc.Theme().CreateLabel()
			label.SetText(item.name)
			table.SetChildAt(0, i, 1, 1, label)

			layout := theme.CreateLinearLayout()
			layout.SetDirection(gxui.LeftToRight)
			layout.SetHorizontalAlignment(gxui.AlignLeft)
			layout.SetVerticalAlignment(gxui.AlignMiddle)

			textbox := theme.CreateTextBox()
			textbox.SetDesiredWidth(math.MaxSize.W)
			textbox.SetText(c.settings[url].(string))
			textbox.OnTextChanged(func([]gxui.TextBoxEdit) {
				c.settings[url] = textbox.Text()
			})
			layout.AddChild(textbox)

			table.SetChildAt(1, i, 1, 1, layout)
		}
		blayout := theme.CreateLinearLayout()
		blayout.SetDirection(gxui.LeftToRight)
		blayout.SetHorizontalAlignment(gxui.AlignLeft)
		blayout.SetVerticalAlignment(gxui.AlignMiddle)
		button := CreateButton(theme, "Update Data")
		button.OnClick(func(gxui.MouseEvent) {
			c.updated = true
			update := new(UpdateDialogContext)
			update.DataLocations = &DataLocations{
				RMDFile:  c.settings["rmd_file"].(string),
				RMDURL:   c.settings["rmd_update_url"].(string),
				APIFile:  c.settings["api_file"].(string),
				APIURL:   c.settings["api_update_url"].(string),
				IconFile: c.settings["icon_file"].(string),
				IconURL:  c.settings["icon_update_url"].(string),
			}
			ctxc.EnterContext(update)
		})
		blayout.AddChild(button)
		table.SetChildAt(0, len(items), 1, 1, blayout)

		layout.AddChild(group("Update URLs", table))
	}

	actions := theme.CreateLinearLayout()
	actions.SetDirection(gxui.LeftToRight)
	actions.SetHorizontalAlignment(gxui.AlignRight)
	actionOK := CreateButton(theme, "OK")
	actionOK.OnClick(func(gxui.MouseEvent) {
		c.ok = true
		ctxc.ExitContext()
	})
	actions.AddChild(actionOK)
	actionCancel := CreateButton(theme, "Cancel")
	actionCancel.OnClick(func(gxui.MouseEvent) {
		c.ok = false
		ctxc.ExitContext()
	})
	actions.AddChild(actionCancel)
	layout.AddChild(actions)

	return []gxui.Control{
		layout,
	}, true
}

func (c *SettingsContext) Exiting(ctxc *ContextController) {
	if c.ok {
		Settings.Sets(c.settings)
	}
	if c.updated {
		Data.Reload(new(DataLocations).FromSettings(Settings))
	}
}

func (c *SettingsContext) IsDialog() bool {
	return false
}

func (c *SettingsContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *SettingsContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignCenter
}

func (c *SettingsContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignTop
}
