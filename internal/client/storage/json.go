package storage

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path"
)

type JSONStorage struct {
	filepath string
	Token    string `json:"auth_token"`
	Login    string `json:"login"`
}

func NewJSONStorage(filepath string) *JSONStorage {
	return &JSONStorage{
		filepath: filepath,
	}
}

func (s *JSONStorage) Load() error {
	storageData, err := os.ReadFile(s.filepath)
	if err == nil && len(storageData) > 0 {
		err = json.Unmarshal(storageData, &s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *JSONStorage) Save() error {
	dir := path.Dir(s.filepath)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return errors.Wrap(err, "Failed to create folder for JSONStorage storage")
	}

	storageData, err := json.Marshal(&s)
	if err != nil {
		return err
	}

	err = os.WriteFile(s.filepath, storageData, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (s *JSONStorage) SetToken(token string) {
	s.Token = token
}

func (s *JSONStorage) GetToken() string {
	return s.Token
}

func (s *JSONStorage) SetLogin(login string) {
	s.Login = login
}

func (s *JSONStorage) GetLogin() string {
	return s.Login
}
