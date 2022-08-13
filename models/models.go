// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import "errors"

// Pagination is used in List repository operations to limit or offset the
// elements returned.
type Pagination struct {
	Limit  int
	Offset int
}

var (
	// Returned when a `Get`, `Update` or `Delete` call matches no entity.
	ErrNotFound          = errors.New("entity not found")
	ErrDuplicateKeyFound = errors.New("duplicate key found")
)
