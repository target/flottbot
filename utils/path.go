package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	errCurrPath      = errors.New("could not get the current directory")
	errPathNotExists = errors.New("directory named %#q does not exist at: %s")
	errPathOther     = errors.New("there was an error attempting to access the directory")
)

// PathExists helps determine whether a path exists
func PathExists(p string) (string, error) {
	ex, err := os.Executable()

	if err != nil {
		return "", errCurrPath
	}

	exPath := filepath.Dir(ex)
	fullPath := filepath.Join(exPath, p)

	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(errPathNotExists.Error(), p, fullPath)
		}
		return "", fmt.Errorf("%s: %v", errPathOther.Error(), err)
	}

	return fullPath, nil
}
