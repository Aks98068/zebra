package output

import (
	"os"
	"path/filepath"
)

func SaveText(filename string, content string) error {

	dir := filepath.Dir(filename)

	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(filename, []byte(content), 0644)
}
