package main

import (
	"fmt"
	"github.com/anaminus/gxui"
	"github.com/anaminus/rbxplore/download"
	"github.com/anaminus/rbxplore/event"
	"github.com/anaminus/rbxplore/settings"
	"github.com/kardianos/osext"
	"github.com/robloxapi/rbxapi"
	"github.com/robloxapi/rbxapi/dump"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/xml"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const RMDFileName = "ReflectionMetadata.xml"
const RMDUpdateURL = "https://anaminus.github.io/rbx/raw/rmd/latest.xml"

const APIFileName = "api.txt"
const APIUpdateURL = "https://anaminus.github.io/rbx/raw/api/latest.txt"

const IconFileName = "icon-explorer.png"
const IconUpdateURL = "http://anaminus.github.io/api/img/icon-explorer.png"

type DataLocations struct {
	RMDFile  string
	RMDURL   string
	APIFile  string
	APIURL   string
	IconFile string
	IconURL  string
}

func (d *DataLocations) FromSettings(s settings.Settings) *DataLocations {
	d.RMDFile = s.Get("rmd_file").(string)
	d.RMDURL = s.Get("rmd_update_url").(string)
	d.APIFile = s.Get("api_file").(string)
	d.APIURL = s.Get("api_update_url").(string)
	d.IconFile = s.Get("icon_file").(string)
	d.IconURL = s.Get("icon_update_url").(string)
	return d
}

type dataStruct struct {
	RMD    *rbxfile.Root
	API    *rbxapi.API
	Icons  map[string]gxui.Texture
	driver gxui.Driver
	dl     *download.Download
}

func (d *dataStruct) OnUpdateProgress(progress func(...interface{})) event.Connection {
	return d.dl.OnProgress(progress)
}

func (d *dataStruct) Reload(l *DataLocations) {
	d.reloadItem("ReflectionMetadata", l.RMDFile, RMDFileName, func(r io.ReadSeeker) error {
		var err error
		d.RMD, err = xml.Deserialize(r, nil)
		if err != nil {
			d.RMD = nil
		}
		return err
	})
	d.reloadItem("API", l.APIFile, APIFileName, func(r io.ReadSeeker) error {
		var err error
		d.API, err = dump.Decode(r)
		if err != nil {
			d.API = nil
		}
		return err
	})
	d.regenerateIcons(l.IconFile, IconFileName)
}

func (d *dataStruct) reloadItem(name, file, fileAlt string, reload func(r io.ReadSeeker) error) {
	file, err := getFileNearExec(file, fileAlt)
	if err != nil {
		goto failed
	}
	{
		f, err := os.Open(file)
		if err != nil {
			goto failed
		}
		err = reload(f)
		f.Close()
		if err != nil {
			goto failed
		}
		return
	}
failed:
	log.Printf("failed to reload %s: %s\n", name, err)
}

func (d *dataStruct) regenerateIcons(file, fileAlt string) {
	if d.driver == nil {
		return
	}

	const size = 16

	var err error
	defer func() {
		if err != nil {
			d.Icons = nil
		}
	}()

	file, _ = getFileNearExec(file, fileAlt)
	f, err := os.Open(file)
	if err != nil {
		log.Println("failed to open icons:", err)
		return
	}
	source, err := png.Decode(f)
	if err != nil {
		log.Println("failed to decode icons:", err)
		f.Close()
		return
	}

	if d.RMD == nil {
		log.Println("failed to generate icons: no ReflectionMetadata")
		return
	}

	// GL uses the pixel array directly, so using SubImage will produce
	// garbled images.
	textures := make([]gxui.Texture, source.Bounds().Max.X/size)
	if len(textures) == 0 {
		return
	}
	for i := range textures {
		rgba := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.Draw(rgba, rgba.Bounds(), source, image.Pt(i*size, 0), draw.Src)
		textures[i] = d.driver.CreateTexture(rgba, 1)
	}

	d.Icons = make(map[string]gxui.Texture, 64)
	d.Icons[""] = textures[0]
	for _, inst := range d.RMD.Instances {
		if inst.ClassName != "ReflectionMetadataClasses" {
			continue
		}
		for _, inst := range inst.Children {
			if inst.ClassName != "ReflectionMetadataClass" {
				continue
			}
			name := inst.Name()
			if name == "" {
				continue
			}
			v := inst.Get("ExplorerImageIndex")
			if v, ok := v.(rbxfile.ValueString); ok {
				index, err := strconv.Atoi(string(v))
				if err != nil {
					continue
				}
				if index >= len(textures) || index < 0 {
					continue
				}
				d.Icons[name] = textures[index]
			}
		}
	}
}

func (d *dataStruct) Update(l *DataLocations) error {
	if err := d.updateItem("ReflectionMetadata", l.RMDFile, RMDFileName, l.RMDURL, RMDUpdateURL); err != nil {
		return err
	}
	if err := d.updateItem("API dump", l.APIFile, APIFileName, l.APIURL, APIUpdateURL); err != nil {
		return err
	}
	if err := d.updateItem("Icons", l.IconFile, IconFileName, l.IconURL, IconUpdateURL); err != nil {
		return err
	}
	return nil
}

func (d *dataStruct) updateItem(name, file, fileAlt, url, urlAlt string) error {
	file, err := getFileNearExec(file, fileAlt)
	if err != nil {
		return fmt.Errorf("update %s: %s", name, err)
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("update %s: could not open file `%s`: %s", name, file, err)
	}
	defer f.Close()

	if url == "" {
		url = urlAlt
	}

	// Cancel a previously running download, if there is one.
	d.dl.Close()

	tmp, err := ioutil.TempFile(os.TempDir(), "rbxplore")
	if err != nil {
		return fmt.Errorf("update %s: could not create temporary file: %s", name, err)
	}
	defer func() {
		tmpName := filepath.Join(os.TempDir(), tmp.Name())
		tmp.Close()
		os.Remove(tmpName)
	}()

	d.dl.Name = name
	d.dl.URL = url
	if _, err := io.Copy(tmp, d.dl); err != nil {
		return fmt.Errorf("failed to update %s: %s", name, err)
	}

	tmp.Seek(0, os.SEEK_SET)
	if _, err := io.Copy(f, tmp); err != nil {
		return fmt.Errorf("failed to write %s: %s", name, err)
	}

	return nil
}

func getFileNearExec(file, alt string) (string, error) {
	if file != "" {
		return file, nil
	}
	exec, err := osext.ExecutableFolder()
	if err != nil {
		return "", fmt.Errorf("could not get location of `%s`: %s", alt, err)
	}
	return filepath.Join(exec, alt), nil
}

func (d *dataStruct) CancelUpdate() error {
	return d.dl.Close()
}

var Data *dataStruct

func InitData(driver gxui.Driver) {
	Data = &dataStruct{
		driver: driver,
		dl: &download.Download{
			UpdateRate: time.Millisecond * 50,
		},
	}
}
