package protocode

import (
	"errors"
)

var (
	ErrImportNotFound  = errors.New("could not find import")
	ErrMessageNotFound = errors.New("could not find message")
)
