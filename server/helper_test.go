package server_test

import (
	"bytes"
	"encoding/json"
	"github.com/nfisher/nraft/state"
	"io"
	"testing"
)

type Assert struct {
	*testing.T
}

func (a *Assert) Int(actual int) *intAssert {
	return &intAssert{
		T:      a.T,
		actual: actual,
	}
}

func (a *Assert) NilError(v error) {
	a.Helper()
	if v != nil {
		a.Errorf("want error=nil, got %v", v)
	}
}

func (a *Assert) IsFalse(actual bool) {
	a.Helper()
	if actual != false {
		a.Errorf("want false, got %v", actual)
	}
}

func (a *Assert) Term(actual state.Term) *uint64Assert {
	return &uint64Assert{
		T:      a.T,
		actual: uint64(actual),
	}
}

func (a *Assert) Uint64(actual uint64) *uint64Assert {
	return &uint64Assert{
		T:      a.T,
		actual: actual,
	}
}

func (a *Assert) IsTrue(actual bool) {
	a.Helper()
	if actual != true {
		a.Errorf("want true, got %v", actual)
	}
}

func (a *Assert) Len(arr []interface{}, expected int) {
	a.Helper()
	if len(arr) != expected {
		a.Errorf("want len=%v, got %v", expected, len(arr))
	}
}

type uint64Assert struct {
	*testing.T
	actual uint64
}

func (a uint64Assert) EqualTo(expected uint64) {
	a.Helper()
	if a.actual != expected {
		a.Errorf("want %v, got %v", expected, a.actual)
	}
}

type intAssert struct {
	*testing.T
	actual int
}

func (a intAssert) EqualTo(expected int) {
	a.Helper()
	if a.actual != expected {
		a.Errorf("want %v, got %v", expected, a.actual)
	}
}

func encode(req interface{}) (*bytes.Buffer, error) {
	buf, err := json.Marshal(&req)
	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(buf), nil
}

func decode(r io.Reader, resp interface{}) error {
	return json.NewDecoder(r).Decode(resp)
}
