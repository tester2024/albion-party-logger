package protocol

import (
	"errors"
)

var (
	NoConstructor          = errors.New("no constructor")
	EncryptionNotSupported = errors.New("encryption not supported")
)
