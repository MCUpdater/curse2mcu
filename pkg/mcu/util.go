package mcu

import (
	"errors"
	"log"
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
