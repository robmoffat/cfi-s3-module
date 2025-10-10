package cloud

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/cucumber/godog"
)

// TestSuite for running cloud tests
type TestSuite struct {
	*CloudWorld
}

// NewTestSuite creates a new test suite
func NewTestSuite() *TestSuite {
	world := NewCloudWorld()
	return &TestSuite{
		CloudWorld: world,
	}
}

// PortTestParams holds the parameters for port testing
type PortTestParams struct {
	PortNumber  string
	HostName    string
	Protocol    string
	ServiceType string
}

// Setup method called before each scenario with provided parameters
func (suite *TestSuite) setupWithParams(params PortTestParams) {
	// Reset the world for each scenario
	suite.CloudWorld = NewCloudWorld()

	// Setup pre-configured variables for @PerPort tests
	suite.Props["portNumber"] = params.PortNumber
	suite.Props["hostName"] = params.HostName
	suite.Props["protocol"] = params.Protocol
	suite.Props["serviceType"] = params.ServiceType
}

// InitializeScenarioWithParams initializes the scenario context with custom parameters
func (suite *TestSuite) InitializeScenarioWithParams(ctx *godog.ScenarioContext, params PortTestParams) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setupWithParams(params)
		return ctx, nil
	})

	// Register all cloud steps (which includes generic steps)
	suite.RegisterSteps(ctx)
}

// generateHTMLReport runs our custom HTML reporter
func generateHTMLReport(t *testing.T, jsonReportPath, htmlReportPath string) {
	t.Logf("Generating HTML report from %s to %s...", jsonReportPath, htmlReportPath)

	// Run the HTML reporter
	cmd := exec.Command("go", "run", "../generic/example/html-reporter/reporter.go", jsonReportPath, htmlReportPath)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Warning: Failed to generate HTML report: %v\nOutput: %s", err, output)
	} else {
		t.Logf("HTML report generated: %s", htmlReportPath)
	}
}

// RunPortTests runs godog tests for a specific port configuration
func RunPortTests(t *testing.T, params PortTestParams, featuresPath, reportPath string) {
	suite := NewTestSuite()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create output file for JSON report (for HTML generation)
	jsonReportPath := reportPath + ".json"
	htmlReportPath := reportPath + ".html"

	jsonFile, err := os.Create(jsonReportPath)
	if err != nil {
		t.Fatalf("Failed to create %s: %v", jsonReportPath, err)
	}
	defer jsonFile.Close()

	opts := godog.Options{
		Format:   "cucumber",
		Output:   jsonFile,
		Paths:    []string{featuresPath},
		TestingT: t,
	}

	status := godog.TestSuite{
		Name: "Cloud Port Features",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			suite.InitializeScenarioWithParams(ctx, params)
		},
		Options: &opts,
	}.Run()

	// Generate HTML report after tests complete
	generateHTMLReport(t, jsonReportPath, htmlReportPath)

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}
