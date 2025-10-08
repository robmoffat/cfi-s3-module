# Go Cucumber Generic Language Patterns

This document describes the main patterns and step definitions available in the Go-based cucumber testing framework used for CCC (Common Cloud Controls) compliance testing.

## Overview

The `generic_steps.go` file provides a comprehensive set of reusable step definitions that create a mini-language for testing various scenarios. This framework is designed to leverage Go's concurrency primitives (goroutines and channels) and supports both synchronous and asynchronous operations, data validation, and complex object interactions.

## Core Concepts

### Variable Resolution System

The framework uses a sophisticated variable resolution system through the `HandleResolve` method:

- **Direct values**: Simple strings are used as-is
- **Variable references**: Variables wrapped in `{}` are resolved from the test context
- **Special literals**: `{null}`, `{true}`, `{false}`, and numeric values in `{}` are converted to their respective Go types
- **JSONPath expressions**: Complex object paths can be accessed using JSONPath syntax within `{}`

### Test Context (PropsWorld)

All test data is stored in a `PropsWorld` struct that maintains state across steps:

- `pw.Props[key]` - stores and retrieves test data using Go's `map[string]interface{}`
- Results from operations are typically stored in `pw.Props["result"]`
- Thread-safe access is provided through `sync.RWMutex` for concurrent operations

## Step Definition Patterns

### 1. Goroutine and Channel Patterns

#### Basic Goroutine Resolution

```gherkin
Then the function "{functionField}" should resolve
```

- Executes Go functions (including those that return channels or use goroutines)
- Supports `func() interface{}` and `func(context.Context) (interface{}, error)` signatures
- Stores the result in `result` field
- Handles panics gracefully and converts them to errors

#### Timeout-Controlled Goroutine Resolution

```gherkin
Then the function "{functionField}" should resolve within "{timeout}" seconds
```

- Same as basic resolution but with a 10-second timeout using `context.WithTimeout`
- Uses `select` statements with timeout channels for non-blocking operations
- Cancels long-running goroutines using `context.CancelFunc`

### 2. Reflection-Based Function Call Patterns

#### Method Calls on Go Structs

```gherkin
When I call "{struct}" with "methodName"
When I call "{struct}" with "methodName" with parameter "{variable}"
When I call "{struct}" with "methodName" with parameters "{param1}" and "{param2}"
When I call "{struct}" with "methodName" with parameters "{param1}" and "string value" and "{param3}"
```

- Uses Go's `reflect` package to call methods on structs dynamically
- Supports 0-3 parameters with automatic type conversion
- Method names must be exported (capitalized) to be accessible
- Results are stored in `result` field
- Handles method not found errors gracefully

#### Direct Function Calls

```gherkin
When I call "{function}"
When I call "{function}" with parameter "{param1}"
When I call "{function}" with parameters "{param1}" and "{param2}"
When I call "{function}" with parameters "{param1}", "{param2}" and "{param3}"
```

- Calls `func` types stored in the test context
- Supports 1-3 parameters with interface{} conversion
- Uses reflection to match parameter types dynamically
- Results are stored in `result` field

### 3. Go Variable Management Patterns

#### Variable Assignment and Aliasing

```gherkin
When I refer to "{sourceField}" as "targetField"
```

- Creates an alias or deep copies a value from one field to another
- Works with all Go types including slices, maps, and structs
- Uses Go's type system to maintain type safety where possible
- Useful for creating readable variable names and intermediate results

### 4. Go Data Structure Validation Patterns

#### Slice of Maps Validation

```gherkin
Then "{fieldName}" is an slice of objects with the following contents
| property1 | property2 | property3 |
| value1    | value2    | value3    |
| value4    | value5    | value6    |
```

- Validates that a `[]interface{}` contains `map[string]interface{}` with specific properties
- Uses Godog's data tables for structured comparison
- Supports JSONPath expressions for nested property access
- Handles Go type assertions safely

#### Slice Length Validation

```gherkin
Then "{fieldName}" is an slice of objects with length "expectedLength"
```

- Validates the length of Go slices (`[]interface{}`, `[]string`, etc.)
- Uses `len()` function for accurate slice length checking

#### String Slice Validation

```gherkin
Then "{fieldName}" is an slice of strings with the following values
| value |
| str1  |
| str2  |
| str3  |
```

- Validates `[]string` or `[]interface{}` containing strings
- Performs type assertions to ensure string types

#### Map Property Validation

```gherkin
Then "{fieldName}" is an object with the following contents
| property1 | property2 | property3 |
| value1    | value2    | value3    |
```

- Validates that a `map[string]interface{}` has specific key-value pairs
- Supports nested map access using JSONPath syntax
- Handles Go's zero values appropriately

### 5. Go Type Assertion Patterns

#### Nil Checks

```gherkin
Then "{fieldName}" is nil
Then "{fieldName}" is not nil
```

- `nil` checks for Go `nil` values (pointers, slices, maps, channels, functions)
- `not nil` validates that a value is not nil
- `undefined` checks if a key exists in the Props map
- Handles Go's typed nil values correctly

#### Boolean Type Assertions

```gherkin
Then "{fieldName}" is true
Then "{fieldName}" is false
```

- Performs type assertion to `bool` type
- Handles interface{} to bool conversion safely

#### Empty Collection Checks

```gherkin
Then "{fieldName}" is empty
```

- Uses Go's `len()` function for slices, maps, strings, and channels
- Validates that collections have length 0
- Supports all Go collection types

#### String Equality Comparison

```gherkin
Then "{fieldName}" is "{expectedValue}"
```

- Converts both values to strings using `fmt.Sprintf("%v", value)`
- Handles Go's string representation of various types
- Supports variable resolution in expected values

#### Go Error Interface Validation

```gherkin
Then "{fieldName}" is an error
Then "{fieldName}" is an error with message "{expectedMessage}"
```

- Validates that a field implements Go's `error` interface
- Uses type assertion `value.(error)` for error checking
- Compares error messages using `Error()` method

### 6. Go Test Setup and Concurrency Patterns

#### Goroutine-Safe Counter Functions

```gherkin
Given "{handlerName}" is a invocation counter into "{counterField}"
```

- Creates a closure that increments a counter using atomic operations
- Thread-safe for concurrent goroutine access
- Useful for tracking function call frequency in concurrent scenarios

#### Function Mock Creation

```gherkin
Given "{functionName}" is a function which returns a value of "{value}"
```

- Creates a `func() interface{}` that returns a specific value
- Supports both synchronous and context-aware function signatures
- Useful for mocking Go interfaces and function types

#### Goroutine Timing Control

```gherkin
Given we wait for a period of "{milliseconds}" ms
```

- Uses `time.Sleep()` to introduce delays in test execution
- Supports millisecond precision timing
- Useful for testing goroutine synchronization and timing-sensitive operations

### 7. Async Task Management Patterns

#### Goroutine Task Execution

```gherkin
When I start async task "{taskName}"
Then I wait for async task "{taskName}" to complete
Then I wait for async task "{taskName}" to complete within "{timeout}" seconds
```

- Manages long-running goroutines with proper lifecycle control
- Uses `context.Context` for cancellation and timeout handling
- Provides task status monitoring and result collection

#### Concurrent Task Coordination

```gherkin
Then all async tasks should complete
Then async task "{taskName}" should be running
Then async task "{taskName}" should be completed
```

- Coordinates multiple goroutines using `sync.WaitGroup` patterns
- Monitors task states using channels and select statements
- Handles task cancellation and cleanup

## Advanced Go Features

#### JSON Schema Validation Setup

```gherkin
Given schemas loaded from "file" into "variableName"
```

- Loads JSON schemas for data validation
- Sets up schema validators compatible with Go's JSON marshaling
- Configures validation for Go structs and maps

#### JSON Schema Validation with Go Types

The framework includes support for JSON Schema validation optimized for Go:

- Automatically loads schemas and validates against Go struct tags
- Uses Go's `encoding/json` package for marshaling/unmarshaling
- Supports Go's type system including custom types and embedded structs
- Schemas can be referenced in data validation steps using `matches_type(variableName)` suffix

### JSONPath with Go Maps and Slices

Complex object navigation is supported through JSONPath optimized for Go data structures:

- Access nested map properties: `{user.profile.name}`
- Slice indexing: `{users[0].name}`
- Conditional selection: `{users[?(@.active)].name}`
- Works with `map[string]interface{}` and `[]interface{}` types

### Go Error Handling and Panic Recovery

All step definitions include comprehensive Go error handling:

- Panics are recovered and converted to errors using `recover()`
- Errors implement Go's `error` interface and are stored in the `result` field
- Goroutine panics are handled gracefully with proper cleanup
- Provides detailed stack traces for debugging

## Go Usage Examples

### HTTP Client Testing with Goroutines

```gherkin
Given "{httpClient}" is a function which returns a value of "{mockResponse}"
When I call "{httpClient}" with parameter "{requestData}"
Then the function "{result}" should resolve within "{10}" seconds
And "{result}" is an object with the following contents
| status | message |
| 200    | success |
```

### Concurrent Counter Validation

```gherkin
Given "{eventHandler}" is a invocation counter into "eventCount"
When I start async task "{task1}"
When I start async task "{task2}"
Then all async tasks should complete
Then "{eventCount}" is "{2}"
```

### Go Struct and Map Validation

```gherkin
Then "{users}" is an slice of objects with the following contents
| name     | active | profile.email        |
| John Doe | true   | john@example.com     |
| Jane Doe | false  | jane@example.com     |
```

### Channel and Goroutine Coordination

```gherkin
When I start async task "{producer}"
And I start async task "{consumer}"
Then I wait for async task "{producer}" to complete within "{5}" seconds
And I wait for async task "{consumer}" to complete within "{5}" seconds
Then "{result}" is not nil
```

## Best Practices for Go Testing

1. **Use Context for Cancellation**: Always use `context.Context` for long-running operations
2. **Handle Goroutine Leaks**: Ensure all goroutines are properly cleaned up
3. **Type Safety**: Leverage Go's type system with proper type assertions
4. **Error Handling**: Use Go's explicit error handling patterns
5. **Concurrency Safety**: Use mutexes and atomic operations for shared state
6. **Resource Cleanup**: Use `defer` statements for proper resource management

This framework provides a powerful foundation for writing comprehensive, readable, and maintainable Cucumber tests that leverage Go's concurrency features and type system.
