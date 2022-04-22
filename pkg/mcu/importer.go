package mcu

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// ImportPackage is the main entry point into the actual process.
// Cobra has already vetted the basics of our parameters, so we can get right to business.
func ImportPackage(in string, out string) error {
	zh, e := zip.OpenReader(in)
	if e != nil {
		return fmt.Errorf("unable to open zip, %w", e)
	}
	defer func() {
		_ = zh.Close()
	}()

	// first, we're going to a quick pass and pull out the manifest before going any further
	var mFile *zip.File
	for _, f := range zh.File {
		if f.Name == "manifest.json" {
			log.Printf("Found manifest file in zip")
			mFile = f
			break
		}
	}
	if mFile == nil {
		log.Printf("Scanned %v files in zip, but did not find a manifest, this is probably not a curseforge pack.\n", len(zh.File))
		return ErrNoManifest
	}

	mh, e := mFile.Open()
	if e != nil {
		log.Printf("Failed to read manifest from zip")
		return e
	}
	defer func() {
		_ = mh.Close()
	}()

	buf, e := ioutil.ReadAll(mh)
	if e != nil {
		log.Printf("Failed to read manifest from zip")
		return e
	} else {
		log.Printf("Read manifest.json from zip")
	}
	var manifest map[string]interface{}
	if e = json.Unmarshal(buf, &manifest); e != nil {
		log.Printf("Manifest does not contain valid json")
		return e
	} else {
		log.Printf("Successfully parsed manifest json")
	}

	return nil
}
