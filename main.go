package main

import (
	"flag"
	"fmt"
	"github.com/anaminus/rbxplore/settings"
	"io"
	"os"
	"strings"

	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/drivers/gl"
	"github.com/anaminus/gxui/math"
	"github.com/anaminus/gxui/themes/dark"
)

const SettingsFileName = "rbxplore-settings.json"

var Settings settings.Settings

func InitSettings() {
	Settings = settings.Create(SettingsFileName, map[string]interface{}{
		"rmd_file":        "",
		"api_file":        "",
		"icon_file":       "",
		"rmd_update_url":  RMDUpdateURL,
		"api_update_url":  APIUpdateURL,
		"icon_update_url": IconUpdateURL,
		"spawn_processes": true,
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
	InitData(nil)

	if Option.UpdateData {
		conn := Data.OnUpdateProgress(func(v ...interface{}) {
			name := v[0].(string)
			progress := v[1].(int64)
			total := v[2].(int64)
			err, _ := v[3].(error)
			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\rDownload of %s completed\n", name)
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "\r")
			} else {
				if total < 0 {
					fmt.Fprintf(os.Stderr, "\rDownloading %s...", name)

				} else {
					fmt.Fprintf(os.Stderr, "\rDownloading %s (%3.2f%%)...", name, float64(progress)/float64(total)*100)
				}
			}
		})
		err := Data.Update(new(DataLocations).FromSettings(Settings))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		conn.Disconnect()
	}

	Data.Reload(new(DataLocations).FromSettings(Settings))

	session, err := NewSession(Option.InputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode input file: ", err)
		return
	}

	if Option.OutputFormat != "" {
		session.Minified = strings.HasSuffix(Option.OutputFormat, "_min")
		session.Format = FormatFromString(strings.TrimSuffix(Option.OutputFormat, "_min"))
	}

	if Option.OutputFile != "" {
		// TODO: Unless format flag is specified, guess format from output
		// extension, falling back to input format if necessary.
		session.File = Option.OutputFile
		if err := session.EncodeFile(); err != nil {
			fmt.Fprintf(os.Stderr, "could not encode output file: ", err)
			return
		}
	}
}

func guiMain(driver gxui.Driver) {
	InitData(driver)

	theme := dark.CreateTheme(driver)
	window := theme.CreateWindow(800, 600, "rbxplore")

	editor := &EditorContext{}
	ctxc, _ := CreateContextController(driver, window, theme, editor)

	startSession := make(chan bool, 1)
	go func() {
		<-startSession
		driver.Call(func() {
			Data.Reload(new(DataLocations).FromSettings(Settings))
			if Option.InputFile != "" {
				editor.ChangeSession(NewSession(Option.InputFile))
			} else if Option.New {
				editor.ChangeSession(NewSession(""))
			}
		})
	}()

	if Option.UpdateData {
		update := new(UpdateDialogContext)
		update.UpdateFinished = func() {
			startSession <- true
		}
		ctxc.EnterContext(update)
	} else {
		startSession <- true
	}

	window.OnClose(driver.Terminate)
	window.SetPadding(math.Spacing{L: 10, T: 10, R: 10, B: 10})
}

func main() {
	flag.BoolVar(&Option.Debug, "debug", false, "Output debug messages to stdout.")
	flag.StringVar(&Option.SettingsFile, "settings", "", "Read and write settings from `file`. If unspecified, 'rbxplore-settings.json' is read/written from the same location as the executable.")
	flag.BoolVar(&Option.UpdateData, "updatedata", false, "Update ReflectionMetadata and API dump files.")
	flag.BoolVar(&Option.Shell, "shell", false, "Runs the program without a GUI.")
	flag.StringVar(&Option.OutputFile, "output", "", "If --shell is true, export the input file to the given location. The format will be detected from the extension.")
	flag.StringVar(&Option.OutputFormat, "format", "", "If --shell is true, export the input file with the given format. This overrides the output file extension. Valid formats are 'rbxl', 'rbxm', 'rbxlx', 'rbxmx', and 'json'. '_min' may be appended to output in a minified format, if applicable.")
	flag.BoolVar(&Option.New, "new", false, "If running with a GUI, force a new session to be opened.")
	flag.Parse()
	Option.InputFile = flag.Arg(0)
	InitDebug()

	InitSettings()
	Settings.SetFile(Option.SettingsFile)
	Settings.Load()
	Settings.Save()

	if Option.Shell {
		shellMain()
	} else {
		gl.StartDriver(guiMain)
	}
}
