package client

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path"
)

type Params struct {
	AuthToken string `json:"auth_token"`
}

func LoadParams(filepath string) (*Params, error) {
	p := &Params{}
	storageData, err := os.ReadFile(filepath)
	if err == nil && len(storageData) > 0 {
		err = json.Unmarshal(storageData, &p)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func SaveParams(p *Params, filepath string) error {
	dir := path.Dir(filepath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.Wrap(err, "Failed to create folder for Params storage")
	}

	storageData, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, storageData, 0666)
	if err != nil {
		return err
	}

	return nil
}
