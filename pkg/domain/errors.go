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
