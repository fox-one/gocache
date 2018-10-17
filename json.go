package gocache

import (
	"encoding/json"
)

type jsonCoder struct{}

func (j jsonCoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonCoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

var Json Coder = &jsonCoder{}
