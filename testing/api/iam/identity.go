package iam

// Identity represents a cloud identity/user with credentials
type Identity struct {
	UserName    string            // Username or principal name
	Provider    string            // Cloud provider (aws, azure, gcp)
	Credentials map[string]string // Provider-specific credentials (access keys, tokens, etc.)
	ARN         string            // Resource identifier (ARN for AWS, Object ID for Azure, etc.)
	Metadata    map[string]string // Additional metadata
}
