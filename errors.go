package rulesengine

import (
	"fmt"
)

const (
	errNumeric  = "invalid numerical value"
	errOperator = "invalid operator"
	errType     = "invalid value type"
)

type (
	Error struct {
		Message string      `json:"message"`
		Value   interface{} `json:"value"`
	}
)

func (e Error) Error() string {
	return fmt.Sprintf("%s: [%v]", e.Message, e.Value)
}

func newError(msg string, val interface{}) error {
	return Error{
		Message: msg,
		Value:   val,
	}
}
