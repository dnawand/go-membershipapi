package domain

import (
	"fmt"
)

type DataNotFoundError struct {
	DataType string
}

func (e *DataNotFoundError) Error() string {
	return fmt.Sprintf("%s data not found", e.DataType)
}
