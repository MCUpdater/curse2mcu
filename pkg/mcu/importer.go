package mcu

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mcupdater/curse2mcu/pkg/mcu/schema"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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
	var manifest schema.CurseManifest
	if e = json.Unmarshal(buf, &manifest); e != nil {
		log.Printf("Manifest does not contain valid json")
		return e
	} else {
		log.Printf("Successfully parsed manifest json")
	}

	// TODO: validate manifest

	// start creating our serverpack in memory
	var server = &schema.ServerType{
		IdAttr:        "curse2mcu",
		NameAttr:      manifest.Name,
		RevisionAttr:  manifest.Version,
		VersionAttr:   manifest.Minecraft.Version,
		MainClassAttr: schema.MainClass,
	}
	var sp = schema.ServerPack{
		VersionAttr: schema.ServerPackSchemaVersion,
		Server: []*schema.ServerType{
			server,
		},
	}

	// set up header
	// NB: atm, we only support forge
	for i, ml := range manifest.Minecraft.ModLoaders {
		mlId := strings.Split(ml.ID, "-")
		if len(mlId) != 2 {
			log.Printf("Failed to parse mod loader %q, skipping", ml.ID)
		} else if mlId[0] != "forge" {
			log.Printf("Got unsupported mod loader type %q, skipping", mlId[0])
		} else {
			if ml.Primary && i != 0 {
				log.Printf("Warning: 'primary' mod loader %q is not listed first", ml.ID)
			}
			// we've got a forge version, make the entry
			loader := &schema.LoaderType{
				TypeAttr:      "Forge",
				VersionAttr:   manifest.Minecraft.Version + "-" + mlId[1],
				LoadOrderAttr: i,
			}
			server.Loader = append(server.Loader, loader)
		}
	}

	// iterate over mods
	log.Printf("Processing %v mods...", len(manifest.Files))
	for _, file := range manifest.Files {
		fmt.Print(".")
		// TODO: divine names and id's from curse
		url, e := GetCurseURL(file.ProjectID, file.FileID)
		if e != nil {
			return e
		}

		modName := strconv.Itoa(file.ProjectID)
		modId := "curse_" + modName
		mod := &schema.ModuleType{
			ModuleGenericType: &schema.ModuleGenericType{
				NameAttr: modName,
				IdAttr:   modId,
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
			},
		}
		server.Module = append(server.Module, mod)
	}
	fmt.Println()

	// TODO: create override zip if necessary
	if manifest.Overrides != "" {
		log.Printf("Warning: 'overrides' detected, but extraction is not yet implemented, so this pack will be incomplete")
	}

	// create output xml file
	sph, e := os.Create(out)
	if e != nil {
		log.Printf("Failed to create output xml at %q\n", out)
		return e
	}
	defer func() {
		_ = sph.Close()
	}()
	if _, e = sph.WriteString(xml.Header); e != nil {
		log.Printf("Failed to write xml header")
		return e
	}

	// dump our in-memory xml to the file
	buf, e = xml.MarshalIndent(sp, "", "\t")
	if e != nil {
		log.Printf("Failed to convert in-memory serverpack to xml")
		return e
	}

	if n, e := sph.WriteString(string(buf)); e != nil {
		log.Printf("Failed to write results to xml file")
		return e
	} else {
		log.Printf("Successfully wrote %v bytes to %q", n, out)
	}

	return nil
}
