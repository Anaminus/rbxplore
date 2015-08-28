package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

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

func FormatFromString(s string) Format {
	switch s {
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
