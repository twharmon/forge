package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

func Mkdir(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("utils.Mkdir: %w", err)
		}
	}
	return nil
}

func WriteFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0755)
}
