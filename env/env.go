package env

import (
	"bytes"
	"errors"
	"fmt"
)

var errUnmarshalEnv = errors.New("can't unmarshal a nil *Level")

type Env int8

const (
	Debug Env = iota - 1
	Product
)

func (e *Env) Set(s string) error {
	return e.UnmarshalText([]byte(s))
}

func (e *Env) UnmarshalText(text []byte) error {
	if e == nil {
		return errUnmarshalEnv
	}
	if !e.unmarshalText(text) && !e.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized env: %q", text)
	}
	return nil
}

func (e *Env) unmarshalText(text []byte) bool {
	switch string(text) {
	case "debug", "DEBUG":
		*e = Debug
	case "product", "PRODUCT", "": // make the zero value useful
		*e = Product
	default:
		return false
	}
	return true
}
