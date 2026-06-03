package gstool

import (
	"errors"
	"fmt"
)

func Error(msg string, params ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, params...))
}
