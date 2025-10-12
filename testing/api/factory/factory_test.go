package factory

import (
	"testing"

	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

func TestNewFactory(t *testing.T) {
	tests := []struct {
		name     string
		provider CloudProvider
		wantErr  bool
	}{
		{
			name:     "AWS factory",
			provider: ProviderAWS,
			wantErr:  false,
		},
		{
			name:     "Azure factory",
			provider: ProviderAzure,
			wantErr:  false,
		},
		{
			name:     "GCP factory",
			provider: ProviderGCP,
			wantErr:  false,
		},
		{
			name:     "Invalid provider",
			provider: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, err := NewFactory(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && factory == nil {
				t.Error("NewFactory() returned nil factory")
			}
			if !tt.wantErr && factory.GetProvider() != tt.provider {
				t.Errorf("GetProvider() = %v, want %v", factory.GetProvider(), tt.provider)
			}
		})
	}
}

func TestAWSFactory(t *testing.T) {
	factory := NewAWSFactory()

	if factory.GetProvider() != ProviderAWS {
		t.Errorf("GetProvider() = %v, want %v", factory.GetProvider(), ProviderAWS)
	}

	// Test that methods return not implemented errors
	_, err := factory.GetServiceAPI("test-service")
	if err == nil {
		t.Error("GetServiceAPI() should return not implemented error")
	}

	identity := &iam.Identity{
		UserName: "test-user",
		Provider: string(ProviderAWS),
	}
	_, err = factory.GetServiceAPIWithIdentity("test-service", identity)
	if err == nil {
		t.Error("GetServiceAPIWithIdentity() should return not implemented error")
	}

	_, err = factory.GetIAMService()
	if err == nil {
		t.Error("GetIAMService() should return not implemented error")
	}
}

func TestAzureFactory(t *testing.T) {
	factory := NewAzureFactory()

	if factory.GetProvider() != ProviderAzure {
		t.Errorf("GetProvider() = %v, want %v", factory.GetProvider(), ProviderAzure)
	}
}

func TestGCPFactory(t *testing.T) {
	factory := NewGCPFactory()

	if factory.GetProvider() != ProviderGCP {
		t.Errorf("GetProvider() = %v, want %v", factory.GetProvider(), ProviderGCP)
	}
}
