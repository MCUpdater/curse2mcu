package mcu

import (
	"archive/zip"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	tmpDir = ".mcutmp"
)

func init() {
	// ensure we have a temp dir to work with
	if s, err := os.Stat(tmpDir); err == nil {
		if !s.IsDir() {
			panic("Non-directory file exists at " + tmpDir)
		}
		return
	} else if errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir(tmpDir, 0755); err != nil {
			panic("Failed to mkdir " + tmpDir)
		}
	} else {
		panic("Unexpected error checking " + tmpDir)
	}
}

func FileExists(name string) bool {
	if s, err := os.Stat(name); err == nil {
		// no error, as long as it's a file and not a directory, we're good
		return !s.IsDir()
	} else if errors.Is(err, os.ErrNotExist) {
		// unambiguous errors are unambiguous
		return false
	} else {
		// this is bad - we got an error, but we don't know what's wrong
		log.Printf("unable to test file %q, err = %v", name, err)
		return false
	}
}

func AddFileToZip(zipWriter *zip.Writer, name string, dir string) error {
	// open our source file
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	// build file header
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(dir, name)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relPath)

	// set correct method, based on filetype
	ext := filepath.Ext(name)
	if ext == ".zip" || ext == ".jar" {
		header.Method = zip.Store
	} else {
		header.Method = zip.Deflate
	}

	// write file entry
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// DownloadFile fetches the contents of a url to a temporary file and returns a
// handle to the file (that should be closed on the receiver's side) as well as
// its length and md5 checksum.
func DownloadFile(url string) (*os.File, int, string, error) {
	resp, e := http.Get(url)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if e != nil {
		log.Printf("Unable to download %q", url)
		return nil, 0, "", e
	} else if buf, e := ioutil.ReadAll(resp.Body); e != nil {
		log.Printf("Unable to read downloaded file")
		return nil, 0, "", e
	} else if fd, e := ioutil.TempFile(tmpDir, "dl"); e != nil {
		log.Printf("Unable to create temporary file")
		return nil, 0, "", e
	} else if bytes, e := fd.Write(buf); e != nil {
		log.Printf("Failed to write temp file")
		return nil, 0, "", e
	} else {
		sum := fmt.Sprintf("%x", md5.Sum(buf))
		return fd, bytes, sum, nil
	}
}
