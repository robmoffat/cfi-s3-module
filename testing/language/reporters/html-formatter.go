package reporters

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

// Attachment represents a file or data attached to the report
type Attachment struct {
	Name      string
	MediaType string
	Data      []byte
}

// HTMLFormatter is a godog formatter that generates HTML reports
type HTMLFormatter struct {
	out   io.Writer
	title string
	stats struct {
		startTime       time.Time
		endTime         time.Time
		totalFeatures   int
		totalScenarios  int
		passedScenarios int
		failedScenarios int
		totalSteps      int
		passedSteps     int
		failedSteps     int
		skippedSteps    int
		undefinedSteps  int
	}
	bodyBuffer      bytes.Buffer
	scenarioOpened  bool
	featureOpened   bool
	stepKeywords    map[string]string // Maps step AST node IDs to their keywords (Given/When/Then/And/But)
	attachments     []Attachment      // Store attachments for current scenario
	scenarioContext interface{}       // Store scenario context to access attachments
	params          *TestParams       // Optional test parameters
}

// Feature captures feature information
func (f *HTMLFormatter) Feature(gd *messages.GherkinDocument, uri string, c []byte) {
	// Close previous feature if one was opened
	if f.featureOpened {
		// Close scenario if open
		if f.scenarioOpened {
			fmt.Fprintf(&f.bodyBuffer, `</div>`)
			f.scenarioOpened = false
		}
		// Close feature divs (inner div and feature div)
		fmt.Fprintf(&f.bodyBuffer, `</div></div>`)
	}

	f.stats.totalFeatures++
	if gd.Feature != nil {
		// Extract step keywords from the Gherkin document
		for _, child := range gd.Feature.Children {
			if child.Scenario != nil {
				for _, step := range child.Scenario.Steps {
					f.stepKeywords[step.Id] = step.Keyword
				}
			}
			if child.Background != nil {
				for _, step := range child.Background.Steps {
					f.stepKeywords[step.Id] = step.Keyword
				}
			}
		}

		fmt.Fprintf(&f.bodyBuffer, `<div class="feature"><div class="feature-header"><strong>Feature:</strong> %s</div><div>`, gd.Feature.Name)
		f.featureOpened = true
	}
}

// Pickle captures pickle information
func (f *HTMLFormatter) Pickle(pickle *messages.Pickle) {
	// Close previous scenario if one was opened
	if f.scenarioOpened {
		fmt.Fprintf(&f.bodyBuffer, `</div>`)
	}

	f.stats.totalScenarios++
	fmt.Fprintf(&f.bodyBuffer, `<div class="scenario"><strong>Scenario:</strong> %s`, pickle.Name)
	f.scenarioOpened = true
}

// TestRunStarted is required by the formatters.Formatter interface
func (f *HTMLFormatter) TestRunStarted() {
	f.stats.startTime = time.Now()
}

// TestRunFinished captures test run completion
func (f *HTMLFormatter) TestRunFinished(msg *messages.TestRunFinished) {
	f.stats.endTime = time.Now()
}

// Summary generates and writes the final HTML report
func (f *HTMLFormatter) Summary() {
	// Set end time if not already set
	if f.stats.endTime.IsZero() {
		f.stats.endTime = time.Now()
	}

	// Close the last scenario if one was opened
	if f.scenarioOpened {
		fmt.Fprintf(&f.bodyBuffer, `</div>`)
	}

	// Close the last feature if one was opened
	if f.featureOpened {
		// Close feature divs (inner div and feature div)
		fmt.Fprintf(&f.bodyBuffer, `</div></div>`)
	}

	// Generate and write HTML
	html := f.generateHTML()
	fmt.Fprint(f.out, html)
}

// Track step start time (stored temporarily as we can't keep state)
var stepStartTime time.Time

// getStepKeyword returns the keyword for a step by looking up its AST node IDs
func (f *HTMLFormatter) getStepKeyword(step *messages.PickleStep) string {
	// Check if we have any AST node IDs for this step
	if len(step.AstNodeIds) > 0 {
		// Try each AST node ID (usually the first one is the step itself)
		for _, astNodeId := range step.AstNodeIds {
			if keyword, exists := f.stepKeywords[astNodeId]; exists {
				return keyword
			}
		}
	}
	// Fallback to a generic keyword if not found
	return "Step"
}

// Defined is required by the formatters.Formatter interface
func (f *HTMLFormatter) Defined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	stepStartTime = time.Now()
}

// Passed is required by the formatters.Formatter interface
func (f *HTMLFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	duration := time.Since(stepStartTime)
	f.stats.totalSteps++
	f.stats.passedSteps++
	keyword := f.getStepKeyword(step)
	argHTML := formatStepArgument(step.Argument)
	fmt.Fprintf(&f.bodyBuffer, `<div class="step passed"><strong>%s</strong> %s<span class="timestamp" style="float: right;">%s</span>%s</div>`,
		keyword, step.Text, formatDuration(duration), argHTML)
}

// Skipped is required by the formatters.Formatter interface
func (f *HTMLFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	duration := time.Since(stepStartTime)
	f.stats.totalSteps++
	f.stats.skippedSteps++
	keyword := f.getStepKeyword(step)
	argHTML := formatStepArgument(step.Argument)
	fmt.Fprintf(&f.bodyBuffer, `<div class="step skipped"><strong>%s</strong> %s<span class="timestamp" style="float: right;">%s</span>%s</div>`,
		keyword, step.Text, formatDuration(duration), argHTML)
}

// Undefined is required by the formatters.Formatter interface
func (f *HTMLFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	duration := time.Since(stepStartTime)
	f.stats.totalSteps++
	f.stats.undefinedSteps++
	keyword := f.getStepKeyword(step)
	argHTML := formatStepArgument(step.Argument)
	fmt.Fprintf(&f.bodyBuffer, `<div class="step undefined"><strong>%s</strong> %s<span class="timestamp" style="float: right;">%s</span>%s</div>`,
		keyword, step.Text, formatDuration(duration), argHTML)
}

// Failed is required by the formatters.Formatter interface
func (f *HTMLFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition, err error) {
	duration := time.Since(stepStartTime)
	f.stats.totalSteps++
	f.stats.failedSteps++
	f.stats.failedScenarios++ // Track failed scenario
	keyword := f.getStepKeyword(step)
	argHTML := formatStepArgument(step.Argument)
	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf(`<div class="error-message">%s</div>`, err.Error())
	}
	fmt.Fprintf(&f.bodyBuffer, `<div class="step failed"><strong>%s</strong> %s<span class="timestamp" style="float: right;">%s</span>%s%s</div>`,
		keyword, step.Text, formatDuration(duration), argHTML, errMsg)
}

// Pending is required by the formatters.Formatter interface
func (f *HTMLFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	duration := time.Since(stepStartTime)
	keyword := f.getStepKeyword(step)
	argHTML := formatStepArgument(step.Argument)
	fmt.Fprintf(&f.bodyBuffer, `<div class="step pending"><strong>%s</strong> %s<span class="timestamp" style="float: right;">%s</span>%s</div>`,
		keyword, step.Text, formatDuration(duration), argHTML)
}

// formatDuration formats a duration to whole numbers (e.g., 3.4ms -> 3ms)
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.Round(time.Nanosecond).String()
	}
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Second).String()
}

// formatStepArgument formats step arguments (data tables and doc strings)
func formatStepArgument(arg *messages.PickleStepArgument) string {
	if arg == nil {
		return ""
	}

	var buf bytes.Buffer

	// Format data table
	if arg.DataTable != nil && len(arg.DataTable.Rows) > 0 {
		buf.WriteString(`<table class="data-table" style="margin: 10px 0; border-collapse: collapse;">`)
		for i, row := range arg.DataTable.Rows {
			buf.WriteString(`<tr>`)
			for _, cell := range row.Cells {
				if i == 0 {
					// Header row
					buf.WriteString(fmt.Sprintf(`<th style="border: 1px solid #ddd; padding: 8px; background: #f5f5f5;">%s</th>`, cell.Value))
				} else {
					buf.WriteString(fmt.Sprintf(`<td style="border: 1px solid #ddd; padding: 8px;">%s</td>`, cell.Value))
				}
			}
			buf.WriteString(`</tr>`)
		}
		buf.WriteString(`</table>`)
	}

	// Format doc string
	if arg.DocString != nil {
		buf.WriteString(fmt.Sprintf(`<pre class="doc-string" style="margin: 10px 0; padding: 10px; background: #f5f5f5; border-left: 4px solid #ddd;">%s</pre>`, arg.DocString.Content))
	}

	return buf.String()
}

// addSpacesToCamelCase converts camel case to space-separated words
// e.g., "PortNumber" -> "Port Number"
func addSpacesToCamelCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}

// formatAttachments renders attachments as HTML
func formatAttachments(attachments []Attachment) string {
	if len(attachments) == 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString(`<div class="attachments" style="margin: 10px 0;">`)
	buf.WriteString(`<strong>📎 Attachments:</strong>`)

	for _, att := range attachments {
		buf.WriteString(`<div class="attachment" style="margin: 10px 0; padding: 10px; background: #f9f9f9; border-left: 4px solid #2196F3;">`)
		buf.WriteString(fmt.Sprintf(`<div style="font-weight: bold; margin-bottom: 5px;">%s</div>`, att.Name))

		// Handle different media types
		if strings.HasPrefix(att.MediaType, "image/") {
			// Embed images as base64
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			buf.WriteString(fmt.Sprintf(`<img src="data:%s;base64,%s" style="max-width: 100%%; border: 1px solid #ddd;" />`, att.MediaType, encoded))
		} else if att.MediaType == "application/json" {
			// Pretty-print JSON in a collapsible section
			buf.WriteString(fmt.Sprintf(`<details style="margin-top: 5px;"><summary style="cursor: pointer; font-weight: bold;">View JSON (%d bytes)</summary><pre style="margin: 5px 0; padding: 10px; background: #fff; border: 1px solid #ddd; overflow-x: auto; max-height: 400px;">%s</pre></details>`, len(att.Data), string(att.Data)))
		} else {
			// For other types, provide download info
			buf.WriteString(fmt.Sprintf(`<div style="color: #666;">Type: %s, Size: %d bytes</div>`, att.MediaType, len(att.Data)))
		}

		buf.WriteString(`</div>`)
	}

	buf.WriteString(`</div>`)
	return buf.String()
}

// generateHTML creates the HTML report using stats and bodyBuffer
func (f *HTMLFormatter) generateHTML() string {
	totalRunTime := f.stats.endTime.Sub(f.stats.startTime)
	passedScenarios := f.stats.totalScenarios - f.stats.failedScenarios

	// Generate test parameters table if params are available
	paramsTable := ""
	if f.params != nil && (f.params.HostName != "" || f.params.UID != "") {
		var tableRows strings.Builder

		// Use reflection to iterate through struct fields
		v := reflect.ValueOf(*f.params)
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)

			// Format the value based on its type
			var valueStr string
			switch value.Kind() {
			case reflect.Slice:
				// Handle slice types (like Labels)
				if value.Len() > 0 {
					items := make([]string, value.Len())
					for j := 0; j < value.Len(); j++ {
						items[j] = fmt.Sprintf("%v", value.Index(j).Interface())
					}
					valueStr = strings.Join(items, ", ")
				} else {
					valueStr = ""
				}
			default:
				valueStr = fmt.Sprintf("%v", value.Interface())
			}

			// Only add non-empty values
			if valueStr != "" {
				tableRows.WriteString(fmt.Sprintf("<tr><th>%s</th><td>%s</td></tr>", field.Name, valueStr))
			}
		}

		if tableRows.Len() > 0 {
			paramsTable = fmt.Sprintf(`
        <div class="test-params">
            <h2>Test Parameters</h2>
            <table class="params-table">
                %s
            </table>
        </div>`, tableRows.String())
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        .summary { background: #e8f5e9; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .test-params { background: #e3f2fd; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .params-table { width: 100%%; border-collapse: collapse; margin-top: 10px; }
        .params-table th { text-align: left; padding: 8px; background: #2196F3; color: white; width: 30%%; }
        .params-table td { padding: 8px; border-bottom: 1px solid #ddd; }
        .feature { margin: 20px 0; border: 1px solid #ddd; border-radius: 5px; }
        .feature-header { background: #2196F3; color: white; padding: 10px; cursor: pointer; }
        .scenario { margin: 10px; padding: 10px; background:rgba(249, 249, 249, 0.41); border-left: 4px solid #2196F3; }
        .step { padding: 5px 10px; margin: 5px 0; font-family: monospace; }
        .passed { background: #c8e6c9; border-left: 4px solid #e7f7e8; }
        .failed { background: #ffcdd2; border-left: 4px solid #f44336; }
        .skipped { background: #fff9c4; border-left: 4px solid #FFC107; }
        .undefined { background: #e0e0e0; border-left: 4px solid #9E9E9E; }
        .error-message { color: #f44336; font-family: monospace; margin: 10px 0; padding: 10px; background: #ffebee; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🥒 %s</h1>
        %s
        <div class="summary">
            <h2>Summary</h2>
            <p>Generated: %s</p>
            <p>Total Run Time: %s</p>
            <p>Features: %d</p>
            <p>Scenarios: %d (✅ %d | ❌ %d)</p>
            <p>Steps: %d (✅ %d | ❌ %d | ⏭️ %d | ❓ %d)</p>
        </div>
        %s
    </div>
</body>
</html>`,
		f.title,
		f.title,
		paramsTable,
		f.stats.startTime.Format("2006-01-02 15:04:05"),
		formatDuration(totalRunTime),
		f.stats.totalFeatures,
		f.stats.totalScenarios,
		passedScenarios,
		f.stats.failedScenarios,
		f.stats.totalSteps,
		f.stats.passedSteps,
		f.stats.failedSteps,
		f.stats.skippedSteps,
		f.stats.undefinedSteps,
		f.bodyBuffer.String(),
	)
}

// NewHTMLFormatterWithParams creates a new HTML formatter with test parameters
func NewHTMLFormatterWithParams(suite string, out io.Writer, params TestParams) formatters.Formatter {
	f := &HTMLFormatter{
		out:          out,
		title:        suite,
		stepKeywords: make(map[string]string),
		params:       &params,
	}
	f.stats.startTime = time.Now()
	return f
}
