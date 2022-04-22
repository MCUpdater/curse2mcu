package mcu

import (
	"archive/zip"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/mcupdater/curse2mcu/pkg/mcu/schema"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

func GetCurseDownloadURL(project, file int) string {
	return fmt.Sprintf("https://addons-ecs.forgesvc.net/api/v2/addon/%d/file/%d/download-url", project, file)
}

func GetCurseURL(project, file int) (string, error) {
	url := GetCurseDownloadURL(project, file)
	resp, e := http.Get(url)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if e != nil {
		log.Printf("Failed to request mod file url from %q", url)
		return "", e
	} else if buf, e := ioutil.ReadAll(resp.Body); e != nil {
		log.Printf("Failed to read result of %q request", url)
		return "", e
	} else {
		return string(buf), nil
	}
}

func GetCurseModule(file schema.CurseFile) (*schema.ModuleType, error) {
	// request the correct url from the forgesvc api endpoint
	url, e := GetCurseURL(file.ProjectID, file.FileID)
	if e != nil {
		return nil, e
	}

	// fetch the file into temporary space
	fd, size, md5, e := DownloadFile(url)
	defer func() {
		if fd != nil {
			fname := fd.Name()
			_ = fd.Close()
			_ = os.Remove(fname)
		}
	}()
	if e != nil {
		return nil, e
	}

	// read the metadata toml fom the downloaded mod
	zh, e := zip.OpenReader(fd.Name())
	if e != nil {
		log.Printf("Unable to open tmp downloaded mod")
		return nil, e
	}
	defer func() {
		_ = zh.Close()
	}()

	var mFile *zip.File
	for _, f := range zh.File {
		if f.Name == "META-INF/mods.toml" {
			mFile = f
			break
		}
	}
	modName := path.Base(url)
	modId := "curse_" + strconv.Itoa(file.ProjectID)
	if mFile == nil {
		log.Printf("Failed to read metadata toml for %v", modName)
		// this is not fatal, just annoying
	} else {
		mh, e := mFile.Open()
		if e != nil {
			log.Printf("Failed to read metadata toml from zip")
			return nil, e
		}
		defer func() {
			_ = mh.Close()
		}()

		buf, e := ioutil.ReadAll(mh)
		if e != nil {
			log.Printf("Failed to read metadata toml from zip")
			return nil, e
		}

		// parse the toml
		var md map[string]interface{}
		_, e = toml.Decode(string(buf), &md)
		if e != nil {
			log.Printf("Failed to parse metadata toml from zip")
			return nil, e
		}
		//	spew.Dump(md)

		if id, ok := md["mods"].([]map[string]interface{})[0]["modId"]; ok {
			modId = id.(string)
		}
		if name, ok := md["mods"].([]map[string]interface{})[0]["displayName"]; ok {
			modName = name.(string)
		}

		// TODO: import metadata fields
		// TODO: parse out dependencies
	}

	// construct the mod entry
	mod := &schema.ModuleType{
		ModuleGenericType: &schema.ModuleGenericType{
			ModPath:  "mods/" + path.Base(url),
			NameAttr: modName,
			IdAttr:   modId,
			SideAttr: "BOTH",
			ModType: &schema.ModType{
				ModEnum: &schema.ModTypeRegular,
			},
			URL: []*schema.URL{
				&schema.URL{
					Value: url,
				},
			},
			Required: &schema.Required{
				Value:         file.Required,
				IsDefaultAttr: true,
			},
			Size: int64(size),
			MD5:  md5,
		},
	}
	return mod, e
}
