package models

import (
	"bytes"
	"encoding/gob"
)

//EncodeFileDataType encodes file name and contents in one byte array
func EncodeFileDataType(data *FileDataType) ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//DecodeFileDataType decodes byte data array to FileDataType
func DecodeFileDataType(b []byte) (*FileDataType, error) {
	data := &FileDataType{}
	encoder := gob.NewDecoder(bytes.NewReader(b))
	err := encoder.Decode(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
