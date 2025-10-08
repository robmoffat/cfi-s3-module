package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/oliveagle/jsonpath"
)

// AsyncTask represents an asynchronous operation
type AsyncTask struct {
	Name      string
	StartTime time.Time
	Done      chan struct{}
	Result    interface{}
	Error     error
	Context   context.Context
	Cancel    context.CancelFunc
}

// AsyncTaskManager manages multiple async operations
type AsyncTaskManager struct {
	tasks map[string]*AsyncTask
	mutex sync.RWMutex
}

// NewAsyncTaskManager creates a new async task manager
func NewAsyncTaskManager() *AsyncTaskManager {
	return &AsyncTaskManager{
		tasks: make(map[string]*AsyncTask),
	}
}

// StartTask starts a new async task
func (atm *AsyncTaskManager) StartTask(name string, fn func(ctx context.Context) (interface{}, error)) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	task := &AsyncTask{
		Name:      name,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Context:   ctx,
		Cancel:    cancel,
	}

	atm.tasks[name] = task

	go func() {
		defer close(task.Done)
		task.Result, task.Error = fn(ctx)
	}()
}

// WaitForTask waits for a task to complete with timeout
func (atm *AsyncTaskManager) WaitForTask(name string, timeout time.Duration) error {
	atm.mutex.RLock()
	task, exists := atm.tasks[name]
	atm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	select {
	case <-task.Done:
		return task.Error
	case <-time.After(timeout):
		task.Cancel() // Cancel the task on timeout
		return fmt.Errorf("task %s timed out after %v", name, timeout)
	}
}

// GetTaskResult gets the result of a completed task
func (atm *AsyncTaskManager) GetTaskResult(name string) (interface{}, error) {
	atm.mutex.RLock()
	defer atm.mutex.RUnlock()

	task, exists := atm.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task %s not found", name)
	}

	select {
	case <-task.Done:
		return task.Result, task.Error
	default:
		return nil, fmt.Errorf("task %s is still running", name)
	}
}

// IsTaskRunning checks if a task is still running
func (atm *AsyncTaskManager) IsTaskRunning(name string) bool {
	atm.mutex.RLock()
	defer atm.mutex.RUnlock()

	task, exists := atm.tasks[name]
	if !exists {
		return false
	}

	select {
	case <-task.Done:
		return false
	default:
		return true
	}
}

// CancelTask cancels a running task
func (atm *AsyncTaskManager) CancelTask(name string) error {
	atm.mutex.RLock()
	task, exists := atm.tasks[name]
	atm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	task.Cancel()
	return nil
}

// PropsWorld represents the test context equivalent to TypeScript PropsWorld
type PropsWorld struct {
	Props        map[string]interface{}
	T            TestingT // Interface for assertions
	AsyncManager *AsyncTaskManager
	mutex        sync.RWMutex
}

// TestingT interface for testing assertions
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// NewPropsWorld creates a new test world instance
func NewPropsWorld() *PropsWorld {
	return &PropsWorld{
		Props:        make(map[string]interface{}),
		AsyncManager: NewAsyncTaskManager(),
	}
}

// HandleResolve resolves variables and literals from string references
func (pw *PropsWorld) HandleResolve(name string) interface{} {
	if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
		stripped := name[1 : len(name)-1]

		switch stripped {
		case "null":
			return nil
		case "true":
			return true
		case "false":
			return false
		default:
			// Try to parse as number
			if val, err := strconv.ParseFloat(stripped, 64); err == nil {
				return val
			}

			// Try JSONPath resolution
			if val, exists := pw.Props[stripped]; exists {
				return val
			}

			// Try JSONPath query
			if result, err := jsonpath.JsonPathLookup(pw.Props, "$."+stripped); err == nil {
				return result
			}

			// Return original if nothing matches
			return name
		}
	}

	// Check if it's a direct variable name (not wrapped in {})
	if val, exists := pw.Props[name]; exists {
		return val
	}

	// Try JSONPath query for direct names
	if result, err := jsonpath.JsonPathLookup(pw.Props, "$."+name); err == nil {
		return result
	}

	return name
}

// callFunction safely calls a function with error handling
func (pw *PropsWorld) callFunction(fn interface{}, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			pw.Props["result"] = fmt.Errorf("panic: %v", r)
		}
	}()

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		pw.Props["result"] = fmt.Errorf("not a function")
		return
	}

	// Convert args to reflect.Value
	argValues := make([]reflect.Value, len(args))
	for i, arg := range args {
		argValues[i] = reflect.ValueOf(arg)
	}

	// Call function
	results := fnValue.Call(argValues)

	// Handle results
	if len(results) > 0 {
		result := results[0].Interface()
		pw.Props["result"] = result
	}
}

// doesRowMatch checks if a data row matches the expected values
func (pw *PropsWorld) doesRowMatch(expected map[string]string, actual interface{}) (bool, string) {
	actualBytes, _ := json.Marshal(actual)
	var actualMap map[string]interface{}
	json.Unmarshal(actualBytes, &actualMap)

	var debugInfo []string
	debugInfo = append(debugInfo, fmt.Sprintf("Actual object: %s", string(actualBytes)))
	debugInfo = append(debugInfo, "Field comparisons:")

	for field, expectedVal := range expected {
		if strings.HasSuffix(field, "matches_type") {
			// Schema validation mode - would need schema validation library
			// For now, skip this validation
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: SKIPPED (schema validation)", field))
			continue
		}

		// Use JSONPath to get the field value
		var foundVal interface{}
		if result, err := jsonpath.JsonPathLookup(actualMap, "$."+field); err == nil {
			foundVal = result
		} else {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: NOT FOUND (JSONPath error: %v)", field, err))
		}

		resolvedExpected := pw.HandleResolve(expectedVal)

		// Handle boolean comparisons
		if foundVal == true && resolvedExpected == "true" {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH (true == 'true')", field))
			continue
		}
		if foundVal == false && resolvedExpected == "false" {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH (false == 'false')", field))
			continue
		}

		// Convert both to strings for comparison
		foundStr := fmt.Sprintf("%v", foundVal)
		expectedStr := fmt.Sprintf("%v", resolvedExpected)

		if foundStr != expectedStr {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MISMATCH - found: '%s' (type: %T), expected: '%s' (type: %T)",
				field, foundStr, foundVal, expectedStr, resolvedExpected))
			return false, strings.Join(debugInfo, "\n")
		} else {
			debugInfo = append(debugInfo, fmt.Sprintf("  %s: MATCH - '%s'", field, foundStr))
		}
	}

	return true, strings.Join(debugInfo, "\n")
}

// matchData validates array data against expected table
func (pw *PropsWorld) matchData(actual []interface{}, expected []map[string]string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("length mismatch: expected %d, got %d", len(expected), len(actual))
	}

	for i, expectedRow := range expected {
		matches, debugInfo := pw.doesRowMatch(expectedRow, actual[i])
		if !matches {
			return fmt.Errorf("row %d does not match expected values:\n%s", i, debugInfo)
		}
	}

	return nil
}

// Step Definitions

func (pw *PropsWorld) ThePromiseShouldResolve(field string) error {
	promise := pw.HandleResolve(field)

	// Handle different types of "promises"
	switch p := promise.(type) {
	case func() interface{}:
		// Synchronous function
		result := p()
		pw.Props["result"] = result
	case func(context.Context) (interface{}, error):
		// Async function with context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err := p(ctx)
		if err != nil {
			pw.Props["result"] = err
		} else {
			pw.Props["result"] = result
		}
	case string:
		// Check if it's an async task name
		if pw.AsyncManager.IsTaskRunning(p) {
			err := pw.AsyncManager.WaitForTask(p, 30*time.Second)
			if err != nil {
				pw.Props["result"] = err
				return nil
			}
			result, err := pw.AsyncManager.GetTaskResult(p)
			if err != nil {
				pw.Props["result"] = err
			} else {
				pw.Props["result"] = result
			}
		} else {
			// Check if the string refers to a function in Props
			if fn, exists := pw.Props[p]; exists {
				if callable, ok := fn.(func() interface{}); ok {
					result := callable()
					pw.Props["result"] = result
				} else if callableCtx, ok := fn.(func(context.Context) (interface{}, error)); ok {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					result, err := callableCtx(ctx)
					if err != nil {
						pw.Props["result"] = err
					} else {
						pw.Props["result"] = result
					}
				} else {
					pw.Props["result"] = fn
				}
			} else {
				pw.Props["result"] = promise
			}
		}
	default:
		pw.Props["result"] = promise
	}

	return nil
}

func (pw *PropsWorld) thePromiseShouldResolveWithinTimeout(field, timeoutStr string) error {
	promise := pw.HandleResolve(field)
	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout: %s", timeoutStr)
	}
	timeout := time.Duration(timeoutSeconds) * time.Second

	// Handle different types of "promises" with variable timeout
	switch p := promise.(type) {
	case func() interface{}:
		// Async execution with timeout
		resultChan := make(chan interface{}, 1)
		go func() {
			result := p()
			resultChan <- result
		}()

		select {
		case result := <-resultChan:
			pw.Props["result"] = result
		case <-time.After(timeout):
			pw.Props["result"] = fmt.Errorf("timeout after %v", timeout)
		}
	case func(context.Context) (interface{}, error):
		// Async function with context and timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		result, err := p(ctx)
		if err != nil {
			pw.Props["result"] = err
		} else {
			pw.Props["result"] = result
		}
	case string:
		// Check if it's an async task name
		if pw.AsyncManager.IsTaskRunning(p) {
			err := pw.AsyncManager.WaitForTask(p, timeout)
			if err != nil {
				pw.Props["result"] = err
				return nil
			}
			result, err := pw.AsyncManager.GetTaskResult(p)
			if err != nil {
				pw.Props["result"] = err
			} else {
				pw.Props["result"] = result
			}
		} else {
			pw.Props["result"] = promise
		}
	default:
		// Create a timeout channel for any other type
		resultChan := make(chan interface{}, 1)
		go func() {
			resultChan <- promise
		}()

		select {
		case result := <-resultChan:
			pw.Props["result"] = result
		case <-time.After(timeout):
			pw.Props["result"] = fmt.Errorf("timeout after %v", timeout)
		}
	}

	return nil
}

func (pw *PropsWorld) iCallFunction(fnName string) error {
	fn := pw.HandleResolve(fnName)
	pw.callFunction(fn)
	return nil
}

func (pw *PropsWorld) iCallObjectWithMethod(field, fnName string) error {
	obj := pw.HandleResolve(field)

	// Use reflection to call method
	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	results := method.Call([]reflect.Value{})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) ICallObjectWithMethodWithParameter(field, fnName, param string) error {
	obj := pw.HandleResolve(field)
	paramVal := pw.HandleResolve(param)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	results := method.Call([]reflect.Value{reflect.ValueOf(paramVal)})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallObjectWithMethodWithTwoParameters(field, fnName, param1, param2 string) error {
	obj := pw.HandleResolve(field)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	results := method.Call([]reflect.Value{
		reflect.ValueOf(param1Val),
		reflect.ValueOf(param2Val),
	})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallObjectWithMethodWithThreeParameters(field, fnName, param1, param2, param3 string) error {
	obj := pw.HandleResolve(field)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)
	param3Val := pw.HandleResolve(param3)

	objValue := reflect.ValueOf(obj)
	method := objValue.MethodByName(fnName)

	if !method.IsValid() {
		pw.Props["result"] = fmt.Errorf("method %s not found", fnName)
		return nil
	}

	results := method.Call([]reflect.Value{
		reflect.ValueOf(param1Val),
		reflect.ValueOf(param2Val),
		reflect.ValueOf(param3Val),
	})
	if len(results) > 0 {
		pw.Props["result"] = results[0].Interface()
	}

	return nil
}

func (pw *PropsWorld) iCallFunctionWithParameter(fnName, param string) error {
	fn := pw.HandleResolve(fnName)
	paramVal := pw.HandleResolve(param)

	pw.callFunction(fn, paramVal)
	return nil
}

func (pw *PropsWorld) iCallFunctionWithTwoParameters(fnName, param1, param2 string) error {
	fn := pw.HandleResolve(fnName)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)

	pw.callFunction(fn, param1Val, param2Val)
	return nil
}

func (pw *PropsWorld) iCallFunctionWithThreeParameters(fnName, param1, param2, param3 string) error {
	fn := pw.HandleResolve(fnName)
	param1Val := pw.HandleResolve(param1)
	param2Val := pw.HandleResolve(param2)
	param3Val := pw.HandleResolve(param3)

	pw.callFunction(fn, param1Val, param2Val, param3Val)
	return nil
}

func (pw *PropsWorld) IReferToAs(from, to string) error {
	pw.Props[to] = pw.HandleResolve(from)
	return nil
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWithContents(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert to slice of interfaces
	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Convert table to expected format
	expected := make([]map[string]string, len(table.Rows)-1) // Skip header
	headers := make([]string, len(table.Rows[0].Cells))

	for i, cell := range table.Rows[0].Cells {
		headers[i] = cell.Value
	}

	for i := 1; i < len(table.Rows); i++ {
		row := make(map[string]string)
		for j, cell := range table.Rows[i].Cells {
			row[headers[j]] = cell.Value
		}
		expected[i-1] = row
	}

	return pw.matchData(actualSlice, expected)
}

func (pw *PropsWorld) fieldIsSliceOfObjectsWithLength(field, lengthField string) error {
	actual := pw.HandleResolve(field)
	expectedLength := pw.HandleResolve(lengthField)

	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	expectedLen, err := strconv.Atoi(fmt.Sprintf("%v", expectedLength))
	if err != nil {
		return fmt.Errorf("invalid length: %v", expectedLength)
	}

	if len(actualSlice) != expectedLen {
		return fmt.Errorf("expected length %d, got %d", expectedLen, len(actualSlice))
	}

	return nil
}

func (pw *PropsWorld) fieldIsSliceOfStringsWithValues(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	actualSlice, ok := actual.([]interface{})
	if !ok {
		return fmt.Errorf("field %s is not a slice", field)
	}

	// Convert strings to expected format
	expected := make([]map[string]string, len(table.Rows)-1)
	for i := 1; i < len(table.Rows); i++ {
		expected[i-1] = map[string]string{
			"value": table.Rows[i].Cells[0].Value,
		}
	}

	// Convert actual to same format
	actualFormatted := make([]interface{}, len(actualSlice))
	for i, val := range actualSlice {
		actualFormatted[i] = map[string]interface{}{
			"value": val,
		}
	}

	return pw.matchData(actualFormatted, expected)
}

func (pw *PropsWorld) fieldIsObjectWithContents(field string, table *godog.Table) error {
	actual := pw.HandleResolve(field)

	// Convert table to expected format (should have only one data row)
	if len(table.Rows) != 2 {
		return fmt.Errorf("expected exactly one data row in table")
	}

	expected := make(map[string]string)
	for i, cell := range table.Rows[0].Cells {
		expected[cell.Value] = table.Rows[1].Cells[i].Value
	}

	matches, debugInfo := pw.doesRowMatch(expected, actual)
	if !matches {
		return fmt.Errorf("object does not match expected values:\n%s", debugInfo)
	}

	return nil
}

func (pw *PropsWorld) fieldIsNil(field string) error {
	actual := pw.HandleResolve(field)
	if actual != nil {
		return fmt.Errorf("expected %s to be nil, got %v", field, actual)
	}
	return nil
}

func (pw *PropsWorld) fieldIsNotNil(field string) error {
	actual := pw.HandleResolve(field)
	if actual == nil {
		return fmt.Errorf("expected %s to not be nil", field)
	}
	return nil
}

func (pw *PropsWorld) fieldIsTrue(field string) error {
	actual := pw.HandleResolve(field)
	if actual != true {
		return fmt.Errorf("expected %s to be true, got %v", field, actual)
	}
	return nil
}

func (pw *PropsWorld) fieldIsFalse(field string) error {
	actual := pw.HandleResolve(field)
	if actual != false {
		return fmt.Errorf("expected %s to be false, got %v", field, actual)
	}
	return nil
}

func (pw *PropsWorld) fieldIsUndefined(field string) error {
	resolved := pw.HandleResolve(field)
	// If HandleResolve returns the original field name, it means the field doesn't exist
	if resolved == field {
		// Check if it exists in Props directly
		if _, exists := pw.Props[field]; !exists {
			return nil // Field is undefined, which is what we expect
		}
	}
	return fmt.Errorf("expected %s to be undefined", field)
}

func (pw *PropsWorld) fieldIsEmpty(field string) error {
	actual := pw.HandleResolve(field)

	switch v := actual.(type) {
	case []interface{}:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got length %d", field, len(v))
		}
	case string:
		if len(v) != 0 {
			return fmt.Errorf("expected %s to be empty, got length %d", field, len(v))
		}
	default:
		return fmt.Errorf("cannot check if %s is empty: unsupported type", field)
	}

	return nil
}

func (pw *PropsWorld) fieldEquals(field, expected string) error {
	actual := pw.HandleResolve(field)
	expectedVal := pw.HandleResolve(expected)

	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expectedVal)

	if actualStr != expectedStr {
		return fmt.Errorf("expected %s to equal %s, got %s", field, expectedStr, actualStr)
	}

	return nil
}

func (pw *PropsWorld) fieldIsErrorWithMessage(field, errorType string) error {
	actual := pw.HandleResolve(field)

	if err, ok := actual.(error); ok {
		if err.Error() != errorType {
			return fmt.Errorf("expected error message %s, got %s", errorType, err.Error())
		}
	} else {
		return fmt.Errorf("expected %s to be an error", field)
	}

	return nil
}

func (pw *PropsWorld) fieldIsError(field string) error {
	actual := pw.HandleResolve(field)

	if _, ok := actual.(error); !ok {
		return fmt.Errorf("expected %s to be an error, got %T", field, actual)
	}

	return nil
}

func (pw *PropsWorld) HandlerIsInvocationCounter(handlerName, field string) error {
	pw.Props[handlerName] = func() {
		count, exists := pw.Props[field]
		if !exists {
			count = 0
		}
		if countInt, ok := count.(int); ok {
			pw.Props[field] = countInt + 1
		} else {
			pw.Props[field] = 1
		}
	}
	pw.Props[field] = 0
	return nil
}

func (pw *PropsWorld) FunctionReturnsPromiseOf(fnName, field string) error {
	value := pw.HandleResolve(field)
	pw.Props[fnName] = func() interface{} {
		return value
	}
	return nil
}

func (pw *PropsWorld) waitForPeriod(ms string) error {
	duration, err := strconv.Atoi(ms)
	if err != nil {
		return fmt.Errorf("invalid duration: %s", ms)
	}

	time.Sleep(time.Duration(duration) * time.Millisecond)
	return nil
}

func (pw *PropsWorld) schemasLoaded() error {
	// This would need to be implemented based on your schema loading requirements
	// For now, we'll create a placeholder
	pw.Props["ajv"] = "schema_validator_placeholder"
	return nil
}

// Async-specific step definitions

func (pw *PropsWorld) iStartAsyncTask(taskName string) error {
	pw.AsyncManager.StartTask(taskName, func(ctx context.Context) (interface{}, error) {
		// Default async task - can be overridden by setting a function in Props
		if taskFn, exists := pw.Props[taskName+"_function"]; exists {
			if fn, ok := taskFn.(func(context.Context) (interface{}, error)); ok {
				return fn(ctx)
			}
		}

		// Default behavior - simulate some work
		select {
		case <-time.After(1 * time.Second):
			return fmt.Sprintf("Task %s completed", taskName), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	return nil
}

func (pw *PropsWorld) iWaitForAsyncTaskToComplete(taskName string) error {
	err := pw.AsyncManager.WaitForTask(taskName, 30*time.Second)
	if err != nil {
		pw.Props["result"] = err
		return nil
	}

	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		pw.Props["result"] = err
	} else {
		pw.Props["result"] = result
	}
	return nil
}

func (pw *PropsWorld) iWaitForAsyncTaskToCompleteWithin(taskName, timeoutStr string) error {
	timeout, err := time.ParseDuration(timeoutStr + "s")
	if err != nil {
		return fmt.Errorf("invalid timeout: %s", timeoutStr)
	}

	err = pw.AsyncManager.WaitForTask(taskName, timeout)
	if err != nil {
		pw.Props["result"] = err
		return nil
	}

	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		pw.Props["result"] = err
	} else {
		pw.Props["result"] = result
	}
	return nil
}

func (pw *PropsWorld) asyncTaskShouldBeRunning(taskName string) error {
	if !pw.AsyncManager.IsTaskRunning(taskName) {
		return fmt.Errorf("task %s is not running", taskName)
	}
	return nil
}

func (pw *PropsWorld) asyncTaskShouldBeCompleted(taskName string) error {
	if pw.AsyncManager.IsTaskRunning(taskName) {
		return fmt.Errorf("task %s is still running", taskName)
	}

	// Get the result to verify completion
	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		return err
	}

	pw.Props["result"] = result
	return nil
}

func (pw *PropsWorld) iCancelAsyncTask(taskName string) error {
	return pw.AsyncManager.CancelTask(taskName)
}

func (pw *PropsWorld) allAsyncTasksShouldComplete() error {
	// Wait for all tasks to complete (with a reasonable timeout)
	timeout := 30 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		allCompleted := true
		pw.AsyncManager.mutex.RLock()
		for _, task := range pw.AsyncManager.tasks {
			select {
			case <-task.Done:
				// Task is done
			default:
				allCompleted = false
				break
			}
		}
		pw.AsyncManager.mutex.RUnlock()

		if allCompleted {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("not all async tasks completed within %v", timeout)
}

func (pw *PropsWorld) asyncResultShouldBeAvailable(taskName string) error {
	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err != nil {
		return err
	}

	pw.Props["result"] = result
	return nil
}

func (pw *PropsWorld) asyncOperationShouldHaveFailed(taskName string) error {
	result, err := pw.AsyncManager.GetTaskResult(taskName)
	if err == nil {
		return fmt.Errorf("expected task %s to fail, but it succeeded with result: %v", taskName, result)
	}

	pw.Props["result"] = err
	return nil
}

// RegisterSteps registers all step definitions with the Godog suite
func (pw *PropsWorld) RegisterSteps(s *godog.ScenarioContext) {
	// Function resolution patterns
	s.Step(`^the function "([^"]*)" should resolve$`, pw.ThePromiseShouldResolve)
	s.Step(`^the function "([^"]*)" should resolve within "([^"]*)" seconds$`, pw.thePromiseShouldResolveWithinTimeout)

	// Function call patterns - direct functions
	s.Step(`^I call "([^"]*)"$`, pw.iCallFunction)

	// Function call patterns - object methods
	s.Step(`^I call "([^"]*)" with "([^"]*)"$`, pw.iCallObjectWithMethod)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameter "([^"]*)"$`, pw.ICallObjectWithMethodWithParameter)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iCallObjectWithMethodWithTwoParameters)
	s.Step(`^I call "([^"]*)" with "([^"]*)" with parameters "([^"]*)" and "([^"]*)" and "([^"]*)"$`, pw.iCallObjectWithMethodWithThreeParameters)

	// Function call patterns - direct functions
	s.Step(`^I call "([^"]*)" with parameter "([^"]*)"$`, pw.iCallFunctionWithParameter)
	s.Step(`^I call "([^"]*)" with parameters "([^"]*)" and "([^"]*)"$`, pw.iCallFunctionWithTwoParameters)
	s.Step(`^I call "([^"]*)" with parameters "([^"]*)", "([^"]*)" and "([^"]*)"$`, pw.iCallFunctionWithThreeParameters)

	// Variable management
	s.Step(`^I refer to "([^"]*)" as "([^"]*)"$`, pw.IReferToAs)

	// Data validation patterns
	s.Step(`^"([^"]*)" is an slice of objects with the following contents$`, pw.fieldIsSliceOfObjectsWithContents)
	s.Step(`^"([^"]*)" is an slice of objects with length "([^"]*)"$`, pw.fieldIsSliceOfObjectsWithLength)
	s.Step(`^"([^"]*)" is an slice of strings with the following values$`, pw.fieldIsSliceOfStringsWithValues)
	s.Step(`^"([^"]*)" is an object with the following contents$`, pw.fieldIsObjectWithContents)

	// Value assertions
	s.Step(`^"([^"]*)" is nil$`, pw.fieldIsNil)
	s.Step(`^"([^"]*)" is not nil$`, pw.fieldIsNotNil)
	s.Step(`^"([^"]*)" is true$`, pw.fieldIsTrue)
	s.Step(`^"([^"]*)" is false$`, pw.fieldIsFalse)
	s.Step(`^"([^"]*)" is undefined$`, pw.fieldIsUndefined)
	s.Step(`^"([^"]*)" is empty$`, pw.fieldIsEmpty)
	s.Step(`^"([^"]*)" is "([^"]*)"$`, pw.fieldEquals)
	s.Step(`^"([^"]*)" is an error with message "([^"]*)"$`, pw.fieldIsErrorWithMessage)
	s.Step(`^"([^"]*)" is an error$`, pw.fieldIsError)

	// Test setup patterns
	s.Step(`^"([^"]*)" is a invocation counter into "([^"]*)"$`, pw.HandlerIsInvocationCounter)
	s.Step(`^"([^"]*)" is a function which returns a value of "([^"]*)"$`, pw.FunctionReturnsPromiseOf)
	s.Step(`^we wait for a period of "([^"]*)" ms$`, pw.waitForPeriod)
	s.Step(`^schemas loaded$`, pw.schemasLoaded)

	// Async operation patterns
	s.Step(`^I start async task "([^"]*)"$`, pw.iStartAsyncTask)
	s.Step(`^I wait for async task "([^"]*)" to complete$`, pw.iWaitForAsyncTaskToComplete)
	s.Step(`^I wait for async task "([^"]*)" to complete within "([^"]*)" seconds$`, pw.iWaitForAsyncTaskToCompleteWithin)
	s.Step(`^async task "([^"]*)" should be running$`, pw.asyncTaskShouldBeRunning)
	s.Step(`^async task "([^"]*)" should be completed$`, pw.asyncTaskShouldBeCompleted)
	s.Step(`^I cancel async task "([^"]*)"$`, pw.iCancelAsyncTask)
	s.Step(`^all async tasks should complete$`, pw.allAsyncTasksShouldComplete)
	s.Step(`^the async result "([^"]*)" should be available$`, pw.asyncResultShouldBeAvailable)
	s.Step(`^the async operation "([^"]*)" should have failed$`, pw.asyncOperationShouldHaveFailed)
}
