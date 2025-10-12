package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/reporters"
)

// buildServiceTagFilter builds the tag expression for filtering tests based on catalog type
func buildServiceTagFilter(catalogType string) string {
	// Start with @PerService requirement
	tags := []string{"@PerService"}

	// Build exclusion list for all other catalog types (but not the current one)
	var exclusions []string
	for _, ct := range inspection.AllCatalogTypes {
		if ct != catalogType {
			exclusions = append(exclusions, "~@"+ct)
		}
	}

	// Combine: @PerService && ~@otherCatalogType1 && ~@otherCatalogType2 ...
	// This means: run @PerService tests that are NOT tagged with other catalog types
	return strings.Join(append(tags, exclusions...), " && ")
}

// setupServiceParams sets up parameters for @PerService tests
func (suite *TestSuite) setupServiceParams(params reporters.TestParams) {
	// Don't reset CloudWorld - just reset Props
	// This ensures step registrations remain valid
	suite.Props = make(map[string]interface{})

	// Setup pre-configured variables for @PerService tests
	suite.Props["hostName"] = params.HostName
	suite.Props["serviceType"] = params.ServiceType // Deprecated, use providerServiceType or catalogType
	suite.Props["providerServiceType"] = params.ProviderServiceType
	suite.Props["catalogType"] = params.CatalogType
}

// InitializeServiceScenario initializes the scenario context for service testing
func (suite *TestSuite) InitializeServiceScenario(ctx *godog.ScenarioContext, params reporters.TestParams) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setupServiceParams(params)
		return ctx, nil
	})

	// Register all cloud steps (which includes generic steps)
	suite.RegisterSteps(ctx)
}

// RunServiceTests runs godog tests for a specific service configuration
func RunServiceTests(t *testing.T, params reporters.TestParams, featuresPath, reportPath string) {
	suite := NewTestSuite()

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(reportPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create HTML output file
	htmlReportPath := reportPath + ".html"
	ocsfReportPath := reportPath + ".ocsf.json"

	// Create formatter factory with test parameters
	factory := reporters.NewFormatterFactory(params)

	// Register formatters
	godog.Format("html", "HTML report", factory.GetHTMLFormatterFunc())
	godog.Format("ocsf", "OCSF report", factory.GetOCSFFormatterFunc())

	// Build tag filter based on catalog type
	tagFilter := buildServiceTagFilter(params.CatalogType)
	t.Logf("Using tag filter: %s", tagFilter)

	// Create report title
	reportTitle := "Service Test Report: " + params.HostName + " (" + params.CatalogType + " / " + params.ProviderServiceType + ")"

	opts := godog.Options{
		Format:   fmt.Sprintf("html:%s,ocsf:%s", htmlReportPath, ocsfReportPath),
		Paths:    []string{featuresPath},
		Tags:     tagFilter,
		TestingT: nil, // Don't use TestingT to allow proper file output
	}

	status := godog.TestSuite{
		Name: reportTitle,
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			suite.InitializeServiceScenario(ctx, params)
		},
		Options: &opts,
	}.Run()

	// Map godog status to testing behavior
	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.FailNow()
	}
}
