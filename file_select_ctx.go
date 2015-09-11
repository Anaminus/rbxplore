package main

import (
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Adapted from github.xom/google/gxui/samples/file_dlg

var LastFileLocation = ""

var (
	fileColor      = gxui.Color{R: 0.7, G: 0.8, B: 1.0, A: 1}
	directoryColor = gxui.Color{R: 0.8, G: 1.0, B: 0.7, A: 1}
)

// filesAdapter is an implementation of the gxui.ListAdapter interface.
// The AdapterItems returned by this adapter are absolute file path strings.
type filesAdapter struct {
	gxui.AdapterBase
	files []string // The absolute file paths
}

// SetFiles assigns the specified list of absolute-path files to this adapter.
func (a *filesAdapter) SetFiles(files []string) {
	a.files = files
	a.DataChanged(false)
}

func (a *filesAdapter) Count() int {
	return len(a.files)
}

func (a *filesAdapter) ItemAt(index int) gxui.AdapterItem {
	return a.files[index]
}

func (a *filesAdapter) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	for i, f := range a.files {
		if f == path {
			return i
		}
	}
	return -1 // Not found
}

func (a *filesAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	path := a.files[index]
	_, name := filepath.Split(path)
	label := theme.CreateLabel()
	label.SetText(name)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		label.SetColor(directoryColor)
	} else {
		label.SetColor(fileColor)
	}
	return label
}

func (a *filesAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 20}
}

// directoryNode implements the gxui.TreeNode interface to represent a
// directory node in a file-system.
type directoryNode struct {
	path    string   // The absolute path of this directory.
	subdirs []string // The absolute paths of all immediate sub-directories.
}

// Count implements gxui.TreeNodeContainer.
func (d directoryNode) Count() int {
	return len(d.subdirs)
}

// NodeAt implements gxui.TreeNodeContainer.
func (d directoryNode) NodeAt(index int) gxui.TreeNode {
	// Return a directory structure populated with the immediate
	// subdirectories at the given path.
	path := d.subdirs[index]
	directory := directoryNode{path: path}
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err == nil && path != subpath {
			if info.IsDir() {
				directory.subdirs = append(directory.subdirs, subpath)
				return filepath.SkipDir
			}
		}
		return nil
	})
	return directory
}

// ItemIndex implements gxui.TreeNodeContainer.
func (d directoryNode) ItemIndex(item gxui.AdapterItem) int {
	path := item.(string)
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		path += string(filepath.Separator)
	}
	for i, subpath := range d.subdirs {
		subpath += string(filepath.Separator)
		if strings.HasPrefix(path, subpath) {
			return i
		}
	}
	return -1
}

// Item implements gxui.TreeNode.
func (d directoryNode) Item() gxui.AdapterItem {
	return d.path
}

// Create implements gxui.TreeNode.
func (d directoryNode) Create(theme gxui.Theme) gxui.Control {
	_, name := filepath.Split(d.path)
	if name == "" {
		name = d.path
	}
	l := theme.CreateLabel()
	l.SetText(name)
	l.SetColor(directoryColor)
	return l
}

// directoryAdapter is an implementation of the gxui.TreeAdapter interface.
// The AdapterItems returned by this adapter are absolute file path strings.
type directoryAdapter struct {
	gxui.AdapterBase
	directoryNode
}

func (a directoryAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: math.MaxSize.W, H: 22}
}

// Override directoryNode.Create so that the full root is shown, unaltered.
func (a directoryAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	l := theme.CreateLabel()
	l.SetText(a.subdirs[index])
	l.SetColor(directoryColor)
	return l
}

type FileSelectContext struct {
	SelectedFile string
	Saving       bool
	Finished     func()
	ctxc         *ContextController
}

func (c *FileSelectContext) Entering(ctxc *ContextController) ([]gxui.Control, bool) {
	theme := ctxc.Theme()

	// fullpath is the textbox at the top of the window holding the current
	// selection's absolute file path.
	fullpath := theme.CreateTextBox()

	roots := []string{}
	if runtime.GOOS == "windows" {
		for drive := 'A'; drive <= 'Z'; drive++ {
			path := string(drive) + ":"
			if _, err := os.Stat(path); err == nil {
				roots = append(roots, path)
			}
		}
	} else {
		filepath.Walk("/", func(subpath string, info os.FileInfo, err error) error {
			if err == nil && "/" != subpath {
				roots = append(roots, subpath)
				if info.IsDir() {
					return filepath.SkipDir
				}
			}
			return nil
		})
	}

	// directories is the Tree of directories on the left of the window.
	// It uses the directoryAdapter to show the entire system's directory
	// hierarchy.
	directories := theme.CreateTree()
	directories.SetAdapter(&directoryAdapter{
		directoryNode: directoryNode{
			subdirs: roots,
		},
	})

	// filesAdapter is the adapter used to show the currently selected directory's
	// content. The adapter has its data changed whenever the selected directory
	// changes.
	filesAdapter := &filesAdapter{}

	// files is the List of files in the selected directory to the right of the
	// window.
	files := theme.CreateList()
	files.SetAdapter(filesAdapter)

	enterDirOrExit := func(path string) {
		fi, err := os.Stat(path)
		if err == nil && fi.IsDir() {
			if directories.Select(path) {
				directories.Show(path)
			}
			return
		}
		if !c.Saving && os.IsNotExist(err) {
			// If we aren't saving, then do nothing if the file does not
			// exist.
			return
		}
		c.SelectedFile = path
		LastFileLocation = filepath.Dir(path)
		ctxc.ExitContext()
	}

	var open gxui.Button
	if c.Saving {
		open = CreateButton(theme, "Save")
	} else {
		open = CreateButton(theme, "Open")
	}
	open.OnClick(func(gxui.MouseEvent) {
		enterDirOrExit(fullpath.Text())
	})

	cancel := CreateButton(theme, "Cancel")
	cancel.OnClick(func(gxui.MouseEvent) {
		c.SelectedFile = ""
		ctxc.ExitContext()
	})

	// If the user hits the enter key while the fullpath control has focus,
	// attempt to select the directory.
	fullpath.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if ev.Key == gxui.KeyEnter || ev.Key == gxui.KeyKpEnter {
			enterDirOrExit(fullpath.Text())
		}
	})

	// When the directory selection changes, update the files list
	directories.OnSelectionChanged(func(item gxui.AdapterItem) {
		dir := item.(string)
		files := []string{}
		filepath.Walk(dir, func(subpath string, info os.FileInfo, err error) error {
			if err == nil && dir != subpath {
				files = append(files, subpath)
				if info.IsDir() {
					return filepath.SkipDir
				}
			}
			return nil
		})
		filesAdapter.SetFiles(files)
		fullpath.SetText(dir)
		c.SelectedFile = dir
	})

	// When the file selection changes, update the fullpath text
	files.OnSelectionChanged(func(item gxui.AdapterItem) {
		fullpath.SetText(item.(string))
		c.SelectedFile = item.(string)
	})

	// When the user double-clicks a directory in the file list, select it in the
	// directories tree view.
	files.OnDoubleClick(func(gxui.MouseEvent) {
		if path, ok := files.Selected().(string); ok {
			enterDirOrExit(path)
		}
	})

	var startDir string
	if c.SelectedFile != "" {
		startDir = filepath.Dir(c.SelectedFile)
	} else if LastFileLocation == "" {
		if cwd, err := os.Getwd(); err == nil {
			startDir = cwd
		}
	} else {
		startDir = LastFileLocation
	}
	if directories.Select(startDir) {
		directories.Show(directories.Selected())
	}

	splitter := theme.CreateSplitterLayout()
	splitter.SetOrientation(gxui.Horizontal)
	splitter.AddChild(directories)
	splitter.AddChild(files)

	buttonLayout := theme.CreateLinearLayout()
	buttonLayout.SetDirection(gxui.LeftToRight)
	buttonLayout.SetVerticalAlignment(gxui.AlignMiddle)
	buttonLayout.SetHorizontalAlignment(gxui.AlignRight)
	buttonLayout.AddChild(fullpath)
	buttonLayout.AddChild(open)
	buttonLayout.AddChild(cancel)

	fullpath.SetDesiredWidth(math.MaxSize.W)

	topLayout := theme.CreateLinearLayout()
	topLayout.SetDirection(gxui.TopToBottom)
	topLayout.AddChild(buttonLayout)
	topLayout.AddChild(splitter)

	btmLayout := theme.CreateLinearLayout()
	btmLayout.SetDirection(gxui.BottomToTop)
	btmLayout.SetHorizontalAlignment(gxui.AlignRight)
	btmLayout.AddChild(topLayout)

	return []gxui.Control{btmLayout}, true
}

func (c *FileSelectContext) Exiting(ctxc *ContextController) {
	if c.Finished != nil {
		ctxc.driver.Call(c.Finished)
	}
}

func (c *FileSelectContext) IsDialog() bool {
	return false
}

func (c *FileSelectContext) Direction() gxui.Direction {
	return gxui.TopToBottom
}

func (c *FileSelectContext) HorizontalAlignment() gxui.HorizontalAlignment {
	return gxui.AlignLeft
}

func (c *FileSelectContext) VerticalAlignment() gxui.VerticalAlignment {
	return gxui.AlignTop
}
