package main

import (
	"encoding/json"
	"errors"
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
	"github.com/anaminus/rbxplore/action"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/bin"
	"github.com/robloxapi/rbxfile/xml"
)

type Format byte

const (
	FormatNone Format = iota
	FormatRBXL
	FormatRBXM
	FormatRBXLX
	FormatRBXMX
	FormatJSON
)

func (f Format) String() string {
	switch f {
	case FormatRBXL:
		return "rbxl"
	case FormatRBXM:
		return "rbxm"
	case FormatRBXLX:
		return "rbxlx"
	case FormatRBXMX:
		return "rbxmx"
	case FormatJSON:
		return "json"
	}
	return ""
}

func FormatFromString(s string) Format {
	switch strings.ToLower(s) {
	case "rbxl":
		return FormatRBXL
	case "rbxm":
		return FormatRBXM
	case "rbxlx":
		return FormatRBXLX
	case "rbxmx":
		return FormatRBXMX
	case "json":
		return FormatJSON
	}
	return FormatNone
}

type FormatAdapter struct {
	gxui.AdapterBase
}

func (a FormatAdapter) Count() int {
	return 6
}

func (a FormatAdapter) ItemAt(index int) gxui.AdapterItem {
	return Format(index)
}

func (a FormatAdapter) ItemIndex(item gxui.AdapterItem) int {
	return int(item.(Format))
}

func (a FormatAdapter) Create(theme gxui.Theme, index int) gxui.Control {
	l := theme.CreateLabel()
	text := Format(index).String()
	if text == "" {
		text = "None"
	}
	l.SetText(text)
	return l
}

func (a FormatAdapter) Size(gxui.Theme) math.Size {
	return math.Size{W: 60, H: 22}
}

type Session struct {
	File     string
	Format   Format
	Minified bool
	Root     *rbxfile.Root
	Action   *action.Controller
}

func NewSession() *Session {
	return &Session{
		Root:   &rbxfile.Root{},
		Action: action.CreateController(20),
	}
}

// If File is defined, determines Format, and decodes the file into Root.
func (s *Session) DecodeFile() error {
	if s == nil {
		return errors.New("no open session")
	}
	f, err := os.Open(s.File)
	if err != nil {
		return err
	}
	defer f.Close()

	var decode func(io.Reader, *rbxdump.API) (*rbxfile.Root, error)

	// Guess format from file extension.
	if format := FormatFromString(filepath.Ext(f.Name())[1:]); format != FormatNone {
		s.Format = format
	}
	// Otherwise, use given format.
	switch s.Format {
	case FormatRBXL:
		decode = bin.DeserializePlace
	case FormatRBXM:
		decode = bin.DeserializeModel
	case FormatRBXLX:
		decode = xml.Deserialize
	case FormatRBXMX:
		decode = xml.Deserialize
	case FormatJSON:
		d := json.NewDecoder(f)
		if err := d.Decode(s.Root); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("unknown format")
	}

	s.Root, err = decode(f, API)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) EncodeFile() error {
	if s.File == "" {
		return errors.New("no file")
	}
	if s.Format == FormatNone {
		return errors.New("no format")
	}

	f, err := os.Create(s.File)
	if err != nil {
		return err
	}
	defer f.Close()

	var encode func(io.Writer, *rbxdump.API, *rbxfile.Root) error
	switch s.Format {
	case FormatRBXL:
		encode = bin.SerializePlace
	case FormatRBXM:
		encode = bin.SerializeModel
	case FormatRBXLX:
		encode = xml.Serialize
	case FormatRBXMX:
		encode = xml.Serialize
	case FormatJSON:
		e := json.NewEncoder(f)
		if err := e.Encode(s.Root); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("unknown format")
	}

	err = encode(f, API, s.Root)
	if err != nil {
		return err
	}
	return nil
}

////////////////

func (s *Session) AddRootInstance(inst *rbxfile.Instance) action.Action {
	return &actionAddRoot{s.Root, inst}
}

type actionAddRoot struct {
	root     *rbxfile.Root
	instance *rbxfile.Instance
}

func (a *actionAddRoot) Setup() error {
	return nil
}

func (a *actionAddRoot) Forward() error {
	a.root.Instances = append(a.root.Instances, a.instance)
	return nil
}

func (a *actionAddRoot) Backward() error {
	a.root.Instances[len(a.root.Instances)-1] = nil
	a.root.Instances = a.root.Instances[:len(a.root.Instances)-1]
	return nil
}

////////////////

func (s *Session) RemoveRootInstance(i int) action.Action {
	return &actionRemoveRoot{root: s.Root, index: i}
}

type actionRemoveRoot struct {
	root     *rbxfile.Root
	index    int
	instance *rbxfile.Instance
}

func (a *actionRemoveRoot) Setup() error {
	if a.index < 0 || a.index >= len(a.root.Instances) {
		return errors.New("index out of range")
	}
	a.instance = a.root.Instances[a.index]
	return nil
}

func (a *actionRemoveRoot) Forward() error {
	copy(a.root.Instances[a.index:], a.root.Instances[a.index+1:])
	a.root.Instances[len(a.root.Instances)-1] = nil
	a.root.Instances = a.root.Instances[:len(a.root.Instances)-1]
	return nil
}

func (a *actionRemoveRoot) Backward() error {
	a.root.Instances = append(a.root.Instances, nil)
	copy(a.root.Instances[a.index+1:], a.root.Instances[a.index:])
	a.root.Instances[a.index] = a.instance
	return nil
}

////////////////

func (s *Session) SetClassName(inst *rbxfile.Instance, className string) action.Action {
	return &actionSetClassName{instance: inst, newClassName: className}
}

type actionSetClassName struct {
	instance     *rbxfile.Instance
	newClassName string
	oldClassName string
}

func (a *actionSetClassName) Setup() error {
	a.oldClassName = a.instance.ClassName
	return nil
}

func (a *actionSetClassName) Forward() error {
	a.instance.ClassName = a.newClassName
	return nil
}

func (a *actionSetClassName) Backward() error {
	a.instance.ClassName = a.oldClassName
	return nil
}

////////////////

func (s *Session) SetReference(inst *rbxfile.Instance, ref []byte) action.Action {
	c := make([]byte, len(ref))
	copy(c, ref)
	return &actionSetReference{instance: inst, newRef: c}
}

type actionSetReference struct {
	instance *rbxfile.Instance
	newRef   []byte
	oldRef   []byte
}

func (a *actionSetReference) Setup() error {
	a.oldRef = make([]byte, len(a.instance.Reference))
	copy(a.oldRef, a.instance.Reference)
	return nil
}

func (a *actionSetReference) Forward() error {
	a.instance.Reference = a.newRef
	return nil
}

func (a *actionSetReference) Backward() error {
	a.instance.Reference = a.oldRef
	return nil
}

////////////////

func (s *Session) SetParent(inst, parent *rbxfile.Instance) action.Action {
	return &actionSetParent{instance: inst, newParent: parent}
}

type actionSetParent struct {
	instance  *rbxfile.Instance
	newParent *rbxfile.Instance
	oldParent *rbxfile.Instance
}

func (a *actionSetParent) Setup() error {
	a.oldParent = a.instance.Parent()
	return nil
}

func (a *actionSetParent) Forward() error {
	return a.instance.SetParent(a.newParent)
}

func (a *actionSetParent) Backward() error {
	return a.instance.SetParent(a.oldParent)
}

////////////////

func (s *Session) SetIsService(inst *rbxfile.Instance, isService bool) action.Action {
	return &actionSetIsService{instance: inst, newIsService: isService}
}

type actionSetIsService struct {
	instance     *rbxfile.Instance
	newIsService bool
	oldIsService bool
}

func (a *actionSetIsService) Setup() error {
	a.oldIsService = a.instance.IsService
	return nil
}

func (a *actionSetIsService) Forward() error {
	a.instance.IsService = a.newIsService
	return nil
}

func (a *actionSetIsService) Backward() error {
	a.instance.IsService = a.oldIsService
	return nil
}

////////////////

func (s *Session) SetProperty(inst *rbxfile.Instance, prop string, value rbxfile.Value) action.Action {
	return &actionSetProperty{instance: inst, prop: prop, newValue: value}
}

type actionSetProperty struct {
	instance *rbxfile.Instance
	prop     string
	newValue rbxfile.Value
	oldValue rbxfile.Value
}

func (a *actionSetProperty) Setup() error {
	if old, ok := a.instance.Properties[a.prop]; ok {
		a.oldValue = old.Copy()
	}
	return nil
}

func (a *actionSetProperty) Forward() error {
	a.instance.Properties[a.prop] = a.newValue
	return nil
}

func (a *actionSetProperty) Backward() error {
	if a.oldValue == nil {
		delete(a.instance.Properties, a.prop)
	} else {
		a.instance.Properties[a.prop] = a.oldValue
	}
	return nil
}

////////////////
