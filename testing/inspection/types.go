package inspection

// TestParams holds the parameters for port / service testing
// This is the single shared structure used by both inspection and reporters
type TestParams struct {
	PortNumber  string   // Leave blank if not applicable (e.g., for services without specific ports)
	HostName    string   // Hostname or endpoint
	Protocol    string   // Protocol (e.g., "tcp", "https")
	ServiceType string   // Type of service (e.g., "s3", "rds", "storage")
	Region      string   // Cloud region
	Provider    string   // Cloud provider ("aws", "azure", "gcp")
	Labels      []string // Tags/labels from the resource
	UID         string   // Unique identifier (ARN, resource ID, etc.)
}
