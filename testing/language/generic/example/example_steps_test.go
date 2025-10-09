package main

import (
	"fmt"

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
	ctx.Step(`^"([^"]*)" is a function which throws an error$`, es.functionThrowsError)
	ctx.Step(`^"([^"]*)" is a string array with colors$`, es.stringArrayWithColors)
	ctx.Step(`^"([^"]*)" is an empty array$`, es.emptyArray)
	ctx.Step(`^"([^"]*)" is an empty string$`, es.emptyString)
}

// Example-specific step definitions
func (es *ExampleSteps) iHaveAnAPIClientConfiguredIn(variableName string) error {
	// Create and configure the API client in the specified variable
	apiClient := &APIClient{baseURL: "https://api.example.com"}
	es.Props[variableName] = apiClient
	return nil
}

func (es *ExampleSteps) iHaveTestDataIn(variableName string) error {
	// Set up test data in the specified variable
	testData := []interface{}{
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
	es.Props[variableName] = testData
	return nil
}

func (es *ExampleSteps) functionThrowsError(functionName string) error {
	// Create a function that returns an error when called
	errorFunc := func() interface{} {
		return fmt.Errorf("something went wrong")
	}
	es.Props[functionName] = errorFunc
	return nil
}

func (es *ExampleSteps) stringArrayWithColors(variableName string) error {
	// Create a string array with color values
	es.Props[variableName] = []interface{}{"red", "blue", "green"}
	return nil
}

func (es *ExampleSteps) emptyArray(variableName string) error {
	// Create an empty array
	es.Props[variableName] = []interface{}{}
	return nil
}

func (es *ExampleSteps) emptyString(variableName string) error {
	// Create an empty string
	es.Props[variableName] = ""
	return nil
}
