package factory

import (
	"fmt"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/generic"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// AWSFactory implements the Factory interface for AWS
type AWSFactory struct {
	region string
	config map[string]interface{}
}

// NewAWSFactory creates a new AWS factory
func NewAWSFactory() *AWSFactory {
	return &AWSFactory{
		region: "us-east-1", // Default region
		config: make(map[string]interface{}),
	}
}

// GetServiceAPI returns a generic service API client for the given AWS service ARN
func (f *AWSFactory) GetServiceAPI(serviceID string) (generic.Service, error) {
	// TODO: Implement AWS service client creation
	return nil, fmt.Errorf("GetServiceAPI not yet implemented for AWS: %s", serviceID)
}

// GetServiceAPIWithIdentity returns a service API client authenticated as the given identity
func (f *AWSFactory) GetServiceAPIWithIdentity(serviceID string, identity *iam.Identity) (generic.Service, error) {
	// TODO: Implement AWS service client creation with specific identity
	if identity.Provider != string(ProviderAWS) {
		return nil, fmt.Errorf("identity is not for AWS provider: %s", identity.Provider)
	}
	return nil, fmt.Errorf("GetServiceAPIWithIdentity not yet implemented for AWS: %s", serviceID)
}

// GetIAMService returns the IAM service for AWS
func (f *AWSFactory) GetIAMService() (iam.IAMService, error) {
	// TODO: Implement AWS IAM service
	return nil, fmt.Errorf("GetIAMService not yet implemented for AWS")
}

// GetProvider returns the cloud provider
func (f *AWSFactory) GetProvider() CloudProvider {
	return ProviderAWS
}

// SetRegion sets the AWS region for this factory
func (f *AWSFactory) SetRegion(region string) {
	f.region = region
}
