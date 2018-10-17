package gocache

import "github.com/vmihailenco/msgpack"

type msgpackCoder struct{}

func (m msgpackCoder) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (m msgpackCoder) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

var Msgpack Coder = &msgpackCoder{}
