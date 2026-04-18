package domain

import "errors"

var ErrOrderNotFound = errors.New("not found")
var ErrOrderAlreadyExists = errors.New("already exists")
