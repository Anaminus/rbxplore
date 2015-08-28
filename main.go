package main

import (
	"flag"
	"fmt"
	"github.com/anaminus/rbxplore/settings"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/math"
	"github.com/google/gxui/themes/dark"
)

const SettingsFileName = "rbxplore-settings.json"

var Settings settings.Settings

func InitSettings() {
	Settings = settings.Create(SettingsFileName, map[string]interface{}{
		"rmd_file":       "",
		"api_file":       "",
		"rmd_update_url": RMDUpdateURL,
		"api_update_url": APIUpdateURL,
	})
}

var Option struct {
	Debug        bool
	SettingsFile string
	UpdateData   bool
	Shell        bool
	OutputFile   string
	OutputFormat string
	New          bool
	InputFile    string
}

func shellMain() {
}

func guiMain(driver gxui.Driver) {
	GenerateIconTextures(driver)

	theme := dark.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "rbxplore")

	editor := &EditorContext{
		session: &Session{},
	}
	CreateContextController(driver, window, theme, editor)

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}

func main() {
	flag.BoolVar(&Option.Debug, "debug", false, "Output debug messages to stdout.")
	InitDebug()

	flag.StringVar(&Option.SettingsFile, "settings", "", "Read and write settings from `file`. If unspecified, 'rbxplore-settings.json' is read/written from the same location as the executable.")
	flag.BoolVar(&Option.UpdateData, "updatedata", false, "Update ReflectionMetadata and API dump files.")
	flag.BoolVar(&Option.Shell, "shell", false, "Runs the program without a GUI.")
	flag.StringVar(&Option.OutputFile, "output", "", "If --shell is true, export the input file to the given location. The format will be detected from the extension.")
	flag.StringVar(&Option.OutputFormat, "format", "", "If --shell is true, export the input file with the given format. This overrides the output file extension. Valid formats are 'rbxl', 'rbxm', 'rbxlx', 'rbxmx', and 'json'. '_min' may be appended to output in a minimized format, if applicable.")
	flag.BoolVar(&Option.New, "new", true, "If running with a GUI, force a new session to be opened.")
	flag.Parse()
	Option.InputFile = flag.Arg(0)

	InitSettings()
	Settings.SetFile(Option.SettingsFile)
	Settings.Load()

	InitData()

	if Option.Shell {
		shellMain()
	} else {
		gl.StartDriver(guiMain)
	}
}
