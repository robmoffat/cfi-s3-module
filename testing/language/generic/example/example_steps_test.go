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

	// Async test helper functions
	ctx.Step(`^"([^"]*)" is a function which adds 10 to a number$`, es.functionAdds10)
	ctx.Step(`^"([^"]*)" is a function which multiplies two numbers$`, es.functionMultiplies)
	ctx.Step(`^"([^"]*)" is a function which concatenates three strings$`, es.functionConcatenates)
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

// Async helper functions
func (es *ExampleSteps) functionAdds10(functionName string) error {
	// Create a function that adds 10 to its parameter
	addFunc := func(numStr string) interface{} {
		num := 0
		fmt.Sscanf(numStr, "%d", &num)
		return fmt.Sprintf("%d", num+10)
	}
	es.Props[functionName] = addFunc
	return nil
}

func (es *ExampleSteps) functionMultiplies(functionName string) error {
	// Create a function that multiplies two numbers
	multiplyFunc := func(a, b string) interface{} {
		numA, numB := 0, 0
		fmt.Sscanf(a, "%d", &numA)
		fmt.Sscanf(b, "%d", &numB)
		return fmt.Sprintf("%d", numA*numB)
	}
	es.Props[functionName] = multiplyFunc
	return nil
}

func (es *ExampleSteps) functionConcatenates(functionName string) error {
	// Create a function that concatenates three strings
	concatFunc := func(a, b, c string) interface{} {
		return a + b + c
	}
	es.Props[functionName] = concatFunc
	return nil
}
