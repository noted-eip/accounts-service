// Package models defines the data types, payloads and repository interfaces
// of the accounts service.
package models

import "errors"

// Pagination is used in List repository operations to limit or offset the
// elements returned.
type Pagination struct {
	Limit  int64
	Offset int64
}

var (
	// Returned when a `Get`, `Update` or `Delete` call matches no entity.
	ErrNotFound = errors.New("entity not found")

	// Returned when a `Create`, try to insert unique value such as email twice.
	ErrDuplicateKeyFound = errors.New("duplicate key found")

	// Returned when a `Update`, try to update non-existant field.
	ErrUpdateInvalidField = errors.New("invalid update field requested")
)
