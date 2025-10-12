package generic

// Bucket represents a storage bucket/container
type Bucket struct {
	ID     string // Unique identifier (name for AWS S3, Azure Storage Account + Container)
	Name   string // Human-readable name
	Region string // Geographic region
}

// Object represents a stored object/blob
type Object struct {
	ID       string   // Unique identifier (key/path)
	BucketID string   // Parent bucket identifier
	Name     string   // Object name/key
	Size     int64    // Size in bytes
	Data     []string // Object content (for small objects)
}

// ObjectStorageService provides operations for object storage testing
// This interface abstracts S3, Azure Blob Storage, and GCS operations
type ObjectStorageService interface {
	Service // Extends the base Service interface

	// Bucket operations
	ListBuckets() ([]Bucket, error)
	CreateBucket(bucketID string) (*Bucket, error)
	DeleteBucket(bucketID string) error

	// Object operations
	ListObjects(bucketID string) ([]Object, error)
	CreateObject(bucketID string, objectID string, data []string) (*Object, error)
	ReadObject(bucketID string, objectID string) (*Object, error)
	DeleteObject(bucketID string, objectID string) error
}
