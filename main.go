package main

import (
	"fmt"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/dark"
)

var tooltips *gxui.ToolTipController

func appMain(driver gxui.Driver) {
	GenerateIconTextures(driver)

	theme := dark.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "RBXplore")

	editor := &EditorContext{
		session: &Session{},
	}
	ctxc, ok := CreateContextController(driver, window, theme, editor)
	fmt.Println("CTXC", ctxc, ok)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}

func main() {
	gl.StartDriver(appMain)
}
