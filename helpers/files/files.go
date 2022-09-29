package files

import (
	"fmt"
	"os"
)

func WriteFile(path string, data ...string) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	for i := range data {
		if _, err = fmt.Fprint(f, data[i]); err != nil {
			return err
		}
	}

	return nil
}
