package mcu

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

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

func GetCurseDownloadURL(project, file int) string {
	return fmt.Sprintf("https://addons-ecs.forgesvc.net/api/v2/addon/%d/file/%d/download-url", project, file)
}

func GetCurseURL(project, file int) (string, error) {
	url := GetCurseDownloadURL(project, file)
	if resp, e := http.Get(url); e != nil {
		log.Printf("Failed to request mod file url from %q", url)
		return "", e
	} else if buf, e := ioutil.ReadAll(resp.Body); e != nil {
		log.Printf("Failed to read result of %q request", url)
		return "", e
	} else {
		return string(buf), nil
	}
}
