package models

import (
	"bytes"
	"encoding/gob"
)

//EncodeLoginDataType encodes login and password in one byte array
func EncodeLoginDataType(data *LoginDataType) ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//DecodeLoginDataType encodes file name and contents in one byte array
func DecodeLoginDataType(b []byte) (*LoginDataType, error) {
	data := &LoginDataType{}
	encoder := gob.NewDecoder(bytes.NewReader(b))
	err := encoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
