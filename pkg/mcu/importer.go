package mcu

import (
	"archive/zip"
	cryptoMd5 "crypto/md5"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mcupdater/curse2mcu/pkg/mcu/schema"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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
	/*
		log.Printf("Processing %v mods...", len(manifest.Files))
		for i, file := range manifest.Files {
			mod, e := GetCurseModule(file)
			if e != nil {
				fmt.Println("!")
				return e
			}
			fmt.Print(".")
			if i > 0 && i%10 == 0 {
				if i%50 == 0 {
					fmt.Println()
				} else {
					fmt.Print(" ")
				}
			}
			server.Module = append(server.Module, mod)
		}
		fmt.Println()
	*/

	// TODO: create override zip if necessary
	if manifest.Overrides != "" {
		or, e := BuildOverrideZip(manifest.Overrides, zh.File)
		if e != nil {
			log.Printf("Warning: Failed to package override zip, serverpack will be incomplete.")
			return e
		} else {
			// compute md5 and size of override zip
			size := int64(0)
			md5 := ""

			orFileName := manifest.Overrides + ".zip"
			stat, err := os.Stat(orFileName)
			if err != nil {
				log.Printf("Failed to stat final zip for size")
			} else {
				size = stat.Size()
			}

			orFileH, err := os.Open(orFileName)
			if err != nil {
				log.Printf("Failed to open final zip for md5")
			} else {
				hash := cryptoMd5.New()
				_, err = io.Copy(hash, orFileH)
				md5 = fmt.Sprintf("%x", hash.Sum(nil))
				_ = orFileH.Close()
			}

			or.Size = size
			or.MD5 = md5
			log.Printf("md5 = %v, size = %d", md5, size)

			log.Printf("Note: You will need to provide this zip file, and edit your xml to provide its correct url.")
			server.Module = append(server.Module, or)
		}
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

func BuildOverrideZip(prefix string, zipFiles []*zip.File) (*schema.ModuleType, error) {
	dir, err := os.MkdirTemp(tmpDir, "or_tmp_*")
	if err != nil {
		log.Printf("Failed to create tempdir to work on override zip.")
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	// extract all override files
	log.Printf("Processing %d potential override files...", len(zipFiles))
	prefixDir := prefix + "/"
	count := 0
	for i, f := range zipFiles {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name, prefixDir) {
			// we have a candidate
			tmpName := strings.TrimPrefix(f.Name, prefixDir)
			tmpFile := filepath.Join(dir, tmpName)
			_ = os.MkdirAll(filepath.Dir(tmpFile), 0755)
			if tmpH, e := os.Create(tmpFile); e != nil {
				fmt.Print("!")
			} else {
				// yes, this is ugly nested, but we can't use defers for cleanup
				// because we're in a loop - could break these out into multiple
				// subroutines, but that feels even dirtier
				r, e := f.Open()
				if e != nil {
					fmt.Print("!")
				} else {
					buf, e := ioutil.ReadAll(r)
					if e != nil {
						fmt.Print("!")
					} else if _, e = tmpH.Write(buf); e != nil {
						fmt.Print("!")
					} else {
						// success!
						fmt.Print("+")
						count++
					}
					_ = r.Close()
				}
				_ = tmpH.Close()
			}
		} else {
			fmt.Print(".")
		}
		if i > 0 && i%10 == 0 {
			if i%50 == 0 {
				fmt.Println()
			} else {
				fmt.Print(" ")
			}
		}
	}
	fmt.Println()

	// zip them all into a new bundle
	orZip, err := os.Create(prefix + ".zip")
	if err != nil {
		log.Printf("Failed to create override zip file")
		return nil, err
	}
	defer func() {
		_ = orZip.Close()
	}()

	zipWriter := zip.NewWriter(orZip)
	defer func() {
		// don't know if we actually need to close this - but it doesn't hurt
		_ = zipWriter.Close()
	}()

	log.Printf("Packaging %d files into %v...", count, orZip.Name())
	tmpAbs, _ := filepath.Abs(dir)
	if err = filepath.WalkDir(tmpAbs,
		func(path string, d fs.DirEntry, e error) error {
			return AddFileToZip(zipWriter, path)
		},
	); err != nil {
		log.Printf("Failed to walk tmp dir")
		return nil, err
	}

	// NB: computing the size and md5 of this zip here does not seem to work.
	//     we get realistic answers - but not accurate ones. this is likely
	//     the result of not closing and re-loading the file, which we will do
	//     outside this function instead.
	/*
		// compute size and m5 of zip
		_ = orZip.Sync()
		_, _ = orZip.Seek(0, 0)

		stat, err := os.Stat(orZip.Name())
		if err != nil {
			log.Printf("Failed to stat final zip")
			return nil, err
		}
		size := stat.Size()

		hash := md52.New()
		_, err = io.Copy(hash, orZip)
		md5 := fmt.Sprintf("%x", hash.Sum(nil))
	*/

	// construct the mod entry
	url, err := filepath.Abs(orZip.Name())
	if err != nil {
		log.Printf("Failed to determine path to override zip?!")
		return nil, err
	} else {
		log.Printf("Wrote override zip: %v", url)
	}
	mod := &schema.ModuleType{
		ModuleGenericType: &schema.ModuleGenericType{
			NameAttr: "Overrides",
			IdAttr:   "overrides",
			SideAttr: "BOTH",
			ModType: &schema.ModType{
				ModEnum:    &schema.ModTypeExtract,
				InRootAttr: true,
			},
			URL: []*schema.URL{
				&schema.URL{
					Value: "file:/" + url,
				},
			},
			Required: &schema.Required{
				Value:         true,
				IsDefaultAttr: true,
			},
		},
	}
	return mod, nil
}
