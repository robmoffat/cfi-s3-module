package cloud

import (
	"testing"

	"github.com/cucumber/godog"
)

// Global function for godog CLI (required by godog)
func InitializeScenario(ctx *godog.ScenarioContext) {
	suite := NewTestSuite()
	params := PortTestParams{
		PortNumber:  "443",
		HostName:    "robmoff.at",
		Protocol:    "https",
		ServiceType: "web",
	}
	suite.InitializeScenarioWithParams(ctx, params)
}

// Test matrix for different protocols over secure and plaintext ports
var portTestMatrix = []struct {
	name        string
	params      PortTestParams
	description string
}{
	// HTTPS - Secure
	{
		name: "HTTPS_Secure",
		params: PortTestParams{
			PortNumber:  "443",
			HostName:    "robmoff.at",
			Protocol:    "https",
			ServiceType: "web",
		},
		description: "HTTPS on port 443 (TLS)",
	},
	// HTTP - Plaintext
	{
		name: "HTTP_Plaintext",
		params: PortTestParams{
			PortNumber:  "80",
			HostName:    "robmoff.at",
			Protocol:    "http",
			ServiceType: "web",
		},
		description: "HTTP on port 80 (plaintext)",
	},
	// SSH - Secure
	{
		name: "SSH_Secure",
		params: PortTestParams{
			PortNumber:  "22",
			HostName:    "robmoff.at",
			Protocol:    "ssh",
			ServiceType: "ssh",
		},
		description: "SSH on port 22 (encrypted)",
	},
	// SMTP - Secure (SMTPS)
	{
		name: "SMTPS_Secure",
		params: PortTestParams{
			PortNumber:  "465",
			HostName:    "secure.emailsrvr.com",
			Protocol:    "smtps",
			ServiceType: "mail",
		},
		description: "SMTPS on port 465 (TLS)",
	},
	// SMTP - Plaintext with STARTTLS
	{
		name: "SMTP_STARTTLS",
		params: PortTestParams{
			PortNumber:  "587",
			HostName:    "secure.emailsrvr.com",
			Protocol:    "smtp",
			ServiceType: "mail",
		},
		description: "SMTP on port 587 (STARTTLS)",
	},
	// FTP - Plaintext
	{
		name: "FTP_Plaintext",
		params: PortTestParams{
			PortNumber:  "21",
			HostName:    "robmoff.at",
			Protocol:    "ftp",
			ServiceType: "file",
		},
		description: "FTP on port 21 (plaintext)",
	},
	// FTPS - Secure
	{
		name: "FTPS_Secure",
		params: PortTestParams{
			PortNumber:  "990",
			HostName:    "robmoff.at",
			Protocol:    "ftps",
			ServiceType: "file",
		},
		description: "FTPS on port 990 (TLS)",
	},
}

// TestCloudPortFeatures tests the C01 features with example parameters
func TestCloudPortFeatures(t *testing.T) {
	for _, tc := range portTestMatrix {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			reportPath := "output/report-" + tc.name
			RunPortTests(t, tc.params, "../../features/CCC.Core/CO1", reportPath)
		})
	}
}
