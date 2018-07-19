package config

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/hashicorp/hcl"
)

func HclUnmarshalDir(dir string, v interface{}) error {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading dir %q: %s", dir, err)
	}

	for _, fi := range fis {
		filepath := path.Join(dir, fi.Name())
		if path.Ext(filepath) != ".hcl" {
			continue
		}
		raw, err := ioutil.ReadFile(filepath)
		if err != nil {
			return fmt.Errorf("could not read %q: %s", fi.Name(), err)
		}
		if err := hcl.Unmarshal(raw, v); err != nil {
			return fmt.Errorf("could not unmarshal hcl: %s", err)
		}
	}
	return nil
}
