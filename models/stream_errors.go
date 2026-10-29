package models

import "errors"

// ErrUnsupportedStreamEvent identifies an event variant this SDK does not support.
var ErrUnsupportedStreamEvent = errors.New("unsupported stream event")
