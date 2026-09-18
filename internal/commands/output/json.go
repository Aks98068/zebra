package output

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveJSON(filename string, data any) error {

	dir := filepath.Dir(filename)

	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}