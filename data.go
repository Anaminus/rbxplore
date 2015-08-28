package main

import (
	"fmt"
	"github.com/kardianos/osext"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anaminus/rbxplore/download"
	"github.com/anaminus/rbxplore/event"
	"github.com/robloxapi/rbxdump"
	"github.com/robloxapi/rbxfile"
	_ "github.com/robloxapi/rbxfile/bin"
	_ "github.com/robloxapi/rbxfile/xml"
)

const RMDFileName = "ReflectionMetadata.xml"
const RMDUpdateURL = "https://anaminus.github.io/rbx/raw/rmd/latest.xml"

const APIFileName = "api.txt"
const APIUpdateURL = "https://anaminus.github.io/rbx/raw/api/latest.txt"

type dataStruct struct {
	RMD *rbxfile.Root
	API *rbxdump.API
	dl  *download.Download
}

func (d *dataStruct) OnUpdateProgress(progress func(...interface{})) *event.Connection {
	return d.dl.OnProgress(progress)
}

func (d *dataStruct) Reload() {
	d.reloadItem("ReflectionMetadata", "rmd_file", RMDFileName, func(r io.ReadSeeker) error {
		var err error
		d.RMD, err = rbxfile.Decode(r)
		if err != nil {
			d.RMD = nil
		}
		return err
	})
	d.reloadItem("API", "api_file", APIFileName, func(r io.ReadSeeker) error {
		var err error
		d.API, err = rbxdump.Decode(r)
		if err != nil {
			d.API = nil
		}
		return err
	})
}

func (d *dataStruct) reloadItem(name, fileSetting, fileAlt string, reload func(r io.ReadSeeker) error) {
	file, err := getFileNearExec(Settings.Get(fileSetting).(string), fileAlt)
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

func (d *dataStruct) Update() error {
	if err := d.updateItem("ReflectionMetadata", "rmd_file", RMDFileName, "rmd_update_url", RMDUpdateURL); err != nil {
		return err
	}
	if err := d.updateItem("API dump", "api_file", APIFileName, "api_update_url", APIUpdateURL); err != nil {
		return err
	}
	return nil
}

func (d *dataStruct) updateItem(name, fileSetting, fileAlt, urlSetting, urlAlt string) error {
	file, _ := Settings.Get(fileSetting).(string)
	file, err := getFileNearExec(file, fileAlt)
	if err != nil {
		return fmt.Errorf("update %s: %s", name, err)
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("update %s: could not open file `%s`: %s", name, file, err)
	}
	defer f.Close()

	url, _ := Settings.Get(urlSetting).(string)
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

var Data *dataStruct

func InitData() {
	Data = &dataStruct{
		dl: &download.Download{
			UpdateRate: time.Millisecond * 50,
		},
	}
}
