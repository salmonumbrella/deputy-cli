package outfmt

import "errors"

// ErrEmptyResult is returned when --fail-empty is set and a list returns no results.
var ErrEmptyResult = errors.New("empty result")
