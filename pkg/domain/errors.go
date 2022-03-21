package domain

import (
	"fmt"
)

var InternalError = fmt.Errorf("interal server error")

type DataNotFoundError struct {
	DataType string
}

func (e *DataNotFoundError) Error() string {
	return fmt.Sprintf("%s data not found", e.DataType)
}
