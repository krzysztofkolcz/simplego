package catalog

import "errors"

var ErrComponentNotFound = errors.New("component not found")
var ErrProductNotFound = errors.New("product not found")
var ErrInvalidName = errors.New("name cannot be empty")
var ErrInvalidCode = errors.New("code cannot be empty")
var ErrInvalidQuantity = errors.New("quantity must be greater than zero")
