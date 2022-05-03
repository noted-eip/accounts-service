package auth

// Role represents the role assigned to the owner of a token.
type Role = string

const (
	// RoleAdmin grants unrestricted access to Noted endpoints and allows
	// bypassing of endpoint rules.
	RoleAdmin Role = "admin"
	// RoleUser is the default set of permissions for Noted. Can be omitted
	// in the token claims.
	RoleUser Role = "user"
)
