package server

import (
	"bytes"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

func EncodeTo(w io.Writer, v interface{}) error {
	return msgpack.NewEncoder(w).Encode(v)
}

func Encode(v interface{}) (*bytes.Buffer, error) {
	buf, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(buf), nil
}

func Decode(r io.Reader, v interface{}) error {
	return msgpack.NewDecoder(r).Decode(v)
}
