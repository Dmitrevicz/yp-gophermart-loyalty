package storage

import "errors"

var ErrNotFound = errors.New("nothing found")

var ErrDuplicateEntry = errors.New("duplicate entry") // or Unique Violation
