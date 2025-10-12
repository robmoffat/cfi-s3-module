package iam

// Identity represents the identity of a user or service principal
// base class that cloud providers can extend to add provider-specific fields
type Identity struct {
	UserName string // Username or principal name
}
