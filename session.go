package main

import (
	"encoding/json"
	"errors"
	"github.com/anaminus/gxui"
	"github.com/anaminus/gxui/math"
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
	switch filepath.Ext(f.Name()) {
	case ".rbxlx":
		s.Format = FormatRBXLX
		decode = xml.Deserialize
	case ".rbxmx":
		s.Format = FormatRBXMX
		decode = xml.Deserialize
	case ".json":
		d := json.NewDecoder(f)
		if err := d.Decode(s.Root); err != nil {
			return err
		}
		return nil
	case ".rbxl":
		s.Format = FormatRBXL
		// Can be either binary or xml.
		format, _ := rbxfile.GuessFormat(f)
		if format == nil {
			return rbxfile.ErrFormat
		}
		f.Seek(0, os.SEEK_SET)
		decode = format.Decode

	case ".rbxm":
		s.Format = FormatRBXM
		// Can be either binary or xml.
		format, _ := rbxfile.GuessFormat(f)
		if format == nil {
			return rbxfile.ErrFormat
		}
		f.Seek(0, os.SEEK_SET)
		decode = format.Decode
	default:
		// Guess format from content.
		format, _ := rbxfile.GuessFormat(f)
		if format == nil {
			return rbxfile.ErrFormat
		}
		f.Seek(0, os.SEEK_SET)
		switch format.Name() {
		case "rbxl":
			s.Format = FormatRBXL
		case "rbxm":
			s.Format = FormatRBXM
		case "rbxlx":
			s.Format = FormatRBXLX
		case "rbxmx":
			s.Format = FormatRBXMX
		}
		decode = format.Decode
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
		return errors.New("bad format")
	}

	err = encode(f, API, s.Root)
	if err != nil {
		return err
	}
	return nil
}
