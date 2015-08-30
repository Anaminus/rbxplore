package main

import (
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/google/gxui"
	"github.com/robloxapi/rbxfile"
)

var IconTextures map[string]gxui.Texture

func GenerateIconTextures(driver gxui.Driver) {
	const size = 16

	var err error
	defer func() {
		if err != nil {
			IconTextures = nil
		}
	}()

	file, _ := getFileNearExec(Settings.Get("icon_file").(string), IconFileName)
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

	f, err = os.Open("ReflectionMetadata.xml")
	if err != nil {
		log.Println("failed to open RMD:", err)
		return
	}
	rmd, err := rbxfile.Decode(f)
	if err != nil {
		log.Println("failed to decode RMD:", err)
		f.Close()
		return
	}
	f.Close()

	// GL uses the pixel array directly, so using SubImage will produce
	// garbled images.
	textures := make([]gxui.Texture, source.Bounds().Max.X/size)
	if len(textures) == 0 {
		return
	}
	for i := range textures {
		rgba := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.Draw(rgba, rgba.Bounds(), source, image.Pt(i*size, 0), draw.Src)
		textures[i] = driver.CreateTexture(rgba, 1)
	}

	IconTextures = make(map[string]gxui.Texture, 64)
	IconTextures[""] = textures[0]
	for _, inst := range rmd.Instances {
		if inst.ClassName != "ReflectionMetadataClasses" {
			continue
		}
		for _, inst := range inst.GetChildren() {
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
				IconTextures[name] = textures[index]
			}
		}
	}
}
