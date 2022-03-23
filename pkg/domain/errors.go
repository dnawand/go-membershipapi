package domain

import (
	"errors"
	"fmt"
)

var ErrInternal = errors.New("interal server error")
var ErrForbidden = errors.New("forbidden")

type ErrDataNotFound struct {
	DataType string
}

func (e *ErrDataNotFound) Error() string {
	return fmt.Sprintf("%s data not found", e.DataType)
}

type ErrInvalidArgument struct {
	Argument string
	Msg      string
}

func (e *ErrInvalidArgument) Error() string {
	if e.Argument != "" {
		return fmt.Sprintf("%s is invalid", e.Msg)
	}
	return e.Msg
}
