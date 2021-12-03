package protocode

import (
	"errors"
)

var (
	ErrRemoteProcedureNotFound = errors.New("could not find remote procedure")
	ErrServiceNotFound         = errors.New("could not find service")
	ErrMessageNotFound         = errors.New("could not find message")
	ErrImportNotFound          = errors.New("could not find import")
	ErrFieldNotFound           = errors.New("could not find field")
	ErrEnumNotFound            = errors.New("could not find enum")
)
