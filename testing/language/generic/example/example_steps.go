package main

import (
	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
)

// ExampleSteps contains example-specific step definitions
type ExampleSteps struct {
	*generic.PropsWorld
}

// NewExampleSteps creates a new example steps instance
func NewExampleSteps(world *generic.PropsWorld) *ExampleSteps {
	return &ExampleSteps{
		PropsWorld: world,
	}
}

// RegisterExampleSteps registers example-specific step definitions
func (es *ExampleSteps) RegisterExampleSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have an API client configured in "([^"]*)"$`, es.iHaveAnAPIClientConfiguredIn)
	ctx.Step(`^I have test data in "([^"]*)"$`, es.iHaveTestDataIn)
}

// Example-specific step definitions
func (es *ExampleSteps) iHaveAnAPIClientConfiguredIn(variableName string) error {
	// This step assumes the API client is already set up in the test suite setup
	// The variableName parameter allows for flexibility in naming
	return nil
}

func (es *ExampleSteps) iHaveTestDataIn(variableName string) error {
	// This step assumes the test data is already set up in the test suite setup
	// The variableName parameter allows for flexibility in naming
	return nil
}
