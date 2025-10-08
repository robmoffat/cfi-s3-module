package main

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
)

// ExampleTestSuite demonstrates how to use the generic steps
type ExampleTestSuite struct {
	*generic.PropsWorld
}

// NewExampleTestSuite creates a new test suite instance
func NewExampleTestSuite() *ExampleTestSuite {
	return &ExampleTestSuite{
		PropsWorld: generic.NewPropsWorld(),
	}
}

// TestingAdapter adapts Go's testing.T to our TestingT interface
type TestingAdapter struct {
	*testing.T
}

func (ta *TestingAdapter) Errorf(format string, args ...interface{}) {
	ta.T.Errorf(format, args...)
}

func (ta *TestingAdapter) FailNow() {
	ta.T.FailNow()
}

// Example API client for testing
type APIClient struct {
	baseURL string
}

func (c *APIClient) Get(endpoint string) map[string]interface{} {
	return map[string]interface{}{
		"status":  200,
		"message": "success",
		"data": []interface{}{
			map[string]interface{}{
				"id":     1,
				"name":   "John Doe",
				"active": true,
			},
			map[string]interface{}{
				"id":     2,
				"name":   "Jane Doe",
				"active": false,
			},
		},
	}
}

func (c *APIClient) Post(endpoint string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status":  201,
		"message": "created",
		"id":      123,
	}
}

// Setup method called before each scenario
func (suite *ExampleTestSuite) setup() {
	// Reset the world for each scenario
	suite.PropsWorld = generic.NewPropsWorld()

	// Setup test data
	suite.Props["apiClient"] = &APIClient{baseURL: "https://api.example.com"}
	suite.Props["testData"] = map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}
	suite.Props["users"] = []interface{}{
		map[string]interface{}{
			"name":   "John Doe",
			"active": true,
			"profile": map[string]interface{}{
				"email": "john@example.com",
			},
		},
		map[string]interface{}{
			"name":   "Jane Doe",
			"active": false,
			"profile": map[string]interface{}{
				"email": "jane@example.com",
			},
		},
	}
}

// InitializeScenario initializes the scenario context
func (suite *ExampleTestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	// Setup before each scenario
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		suite.setup()
		return ctx, nil
	})

	// Register all generic steps
	suite.RegisterSteps(ctx)

	// Register any custom steps specific to this test suite
	ctx.Step(`^I have an API client configured$`, suite.iHaveAnAPIClientConfigured)
	ctx.Step(`^I make a GET request to "([^"]*)"$`, suite.iMakeAGETRequestTo)
}

// Custom step definitions for this example
func (suite *ExampleTestSuite) iHaveAnAPIClientConfigured() error {
	// API client is already set up in setup()
	return nil
}

func (suite *ExampleTestSuite) iMakeAGETRequestTo(endpoint string) error {
	client := suite.Props["apiClient"].(*APIClient)
	result := client.Get(endpoint)
	suite.Props["apiResponse"] = result
	return nil
}

// TestFeatures runs the godog tests
func TestFeatures(t *testing.T) {
	suite := NewExampleTestSuite()
	suite.T = &TestingAdapter{T: t}

	opts := godog.Options{
		Format:   "pretty",
		Paths:    []string{"features"},
		TestingT: t,
	}

	status := godog.TestSuite{
		Name:                "CCC Compliance Tests",
		ScenarioInitializer: suite.InitializeScenario,
		Options:             &opts,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

// Example usage in a regular Go test - simplified version
func TestGenericStepsBasic(t *testing.T) {
	suite := &ExampleTestSuite{
		PropsWorld: generic.NewPropsWorld(),
	}
	suite.T = &TestingAdapter{T: t}
	suite.setup()

	// Basic test to ensure the suite is working
	if suite.Props == nil {
		t.Error("Expected Props to be initialized")
	}

	if suite.AsyncManager == nil {
		t.Error("Expected AsyncManager to be initialized")
	}

	// Test that we can store and retrieve values
	suite.Props["testValue"] = "hello world"
	if suite.Props["testValue"] != "hello world" {
		t.Error("Failed to store and retrieve value")
	}

	t.Log("Basic generic steps test passed")
}
