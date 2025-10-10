package htmlreporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
)

// HTMLFormatter is a godog formatter that generates HTML reports
type HTMLFormatter struct {
	out        io.Writer
	features   []*messages.GherkinDocument
	pickles    []*messages.Pickle
	testRuns   []*messages.TestRunFinished
	stepStatus map[string]map[string]string        // pickle ID -> step ID -> status
	stepTiming map[string]map[string]time.Duration // pickle ID -> step ID -> duration
	stepStart  map[string]map[string]time.Time     // pickle ID -> step ID -> start time
}

// Feature captures feature information
func (f *HTMLFormatter) Feature(gd *messages.GherkinDocument, uri string, c []byte) {
	f.features = append(f.features, gd)
}

// Pickle captures pickle information
func (f *HTMLFormatter) Pickle(pickle *messages.Pickle) {
	f.pickles = append(f.pickles, pickle)
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	if f.stepStart[pickle.Id] == nil {
		f.stepStart[pickle.Id] = make(map[string]time.Time)
	}
}

// TestRunStarted is required by the formatters.Formatter interface
func (f *HTMLFormatter) TestRunStarted() {
	// Not needed for HTML output
}

// TestRunFinished captures test run completion
func (f *HTMLFormatter) TestRunFinished(msg *messages.TestRunFinished) {
	f.testRuns = append(f.testRuns, msg)
}

// Summary generates and writes the final HTML report
func (f *HTMLFormatter) Summary() {
	// This is called at the end - generate HTML here
	html := f.generateHTML()
	fmt.Fprint(f.out, html)
}

// Defined is required by the formatters.Formatter interface
func (f *HTMLFormatter) Defined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepStart[pickle.Id] == nil {
		f.stepStart[pickle.Id] = make(map[string]time.Time)
	}
	f.stepStatus[pickle.Id][step.Id] = "defined"
	f.stepStart[pickle.Id][step.Id] = time.Now()
}

// Passed is required by the formatters.Formatter interface
func (f *HTMLFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	f.stepStatus[pickle.Id][step.Id] = "passed"
	if startTime, ok := f.stepStart[pickle.Id][step.Id]; ok {
		f.stepTiming[pickle.Id][step.Id] = time.Since(startTime)
	}
}

// Skipped is required by the formatters.Formatter interface
func (f *HTMLFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	f.stepStatus[pickle.Id][step.Id] = "skipped"
	if startTime, ok := f.stepStart[pickle.Id][step.Id]; ok {
		f.stepTiming[pickle.Id][step.Id] = time.Since(startTime)
	}
}

// Undefined is required by the formatters.Formatter interface
func (f *HTMLFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	f.stepStatus[pickle.Id][step.Id] = "undefined"
	if startTime, ok := f.stepStart[pickle.Id][step.Id]; ok {
		f.stepTiming[pickle.Id][step.Id] = time.Since(startTime)
	}
}

// Failed is required by the formatters.Formatter interface
func (f *HTMLFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition, err error) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	f.stepStatus[pickle.Id][step.Id] = "failed"
	if startTime, ok := f.stepStart[pickle.Id][step.Id]; ok {
		f.stepTiming[pickle.Id][step.Id] = time.Since(startTime)
	}
}

// Pending is required by the formatters.Formatter interface
func (f *HTMLFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, def *formatters.StepDefinition) {
	if f.stepStatus[pickle.Id] == nil {
		f.stepStatus[pickle.Id] = make(map[string]string)
	}
	if f.stepTiming[pickle.Id] == nil {
		f.stepTiming[pickle.Id] = make(map[string]time.Duration)
	}
	f.stepStatus[pickle.Id][step.Id] = "pending"
	if startTime, ok := f.stepStart[pickle.Id][step.Id]; ok {
		f.stepTiming[pickle.Id][step.Id] = time.Since(startTime)
	}
}

// generateHTML creates the HTML report
func (f *HTMLFormatter) generateHTML() string {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Cucumber Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 3px solid #4CAF50; padding-bottom: 10px; }
        .summary { background: #e8f5e9; padding: 15px; margin: 20px 0; border-radius: 5px; }
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
        <h1>🥒 Cucumber Test Report</h1>
        <div class="summary">
            <h2>Summary</h2>
            <p>Generated: {{ .Timestamp }}</p>
            <p>Features: {{ .TotalFeatures }}</p>
            <p>Scenarios: {{ .TotalScenarios }} (✅ {{ .PassedScenarios }} | ❌ {{ .FailedScenarios }})</p>
        </div>
        {{ range .Features }}
        <div class="feature">
            <div class="feature-header">
                <strong>Feature:</strong> {{ .Name }}
            </div>
            <div>
                {{ range .Scenarios }}
                <div class="scenario">
                    <strong>{{ .Keyword }}:</strong> {{ .Name }}
                    <div class="timestamp">Duration: {{ .Duration }}</div>
                    {{ range .Steps }}
                    <div class="step {{ .Status }}">
                        {{ .Name }}
                        <span class="timestamp" style="float: right;">{{ .Duration }}</span>
                        {{ if .ErrorMessage }}
                        <div class="error-message">{{ .ErrorMessage }}</div>
                        {{ end }}
                    </div>
                    {{ end }}
                </div>
                {{ end }}
            </div>
        </div>
        {{ end }}
    </div>
</body>
</html>`

	data := struct {
		Timestamp       string
		TotalFeatures   int
		TotalScenarios  int
		PassedScenarios int
		FailedScenarios int
		Features        []FeatureData
	}{
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
		TotalFeatures:   len(f.features),
		TotalScenarios:  0,
		PassedScenarios: 0,
		FailedScenarios: 0,
		Features:        []FeatureData{},
	}

	// Build feature map from gherkin documents
	featureMap := make(map[string]*messages.GherkinDocument)
	for _, feature := range f.features {
		if feature.Feature != nil {
			featureMap[feature.Uri] = feature
		}
	}

	// Build feature data from pickles
	featureDataMap := make(map[string]*FeatureData)

	for _, pickle := range f.pickles {
		data.TotalScenarios++

		// Get or create feature data
		featureData, exists := featureDataMap[pickle.Uri]
		if !exists {
			gherkinDoc := featureMap[pickle.Uri]
			featureName := pickle.Uri
			if gherkinDoc != nil && gherkinDoc.Feature != nil {
				featureName = gherkinDoc.Feature.Name
			}
			featureData = &FeatureData{
				Name:      featureName,
				Scenarios: []ScenarioData{},
			}
			featureDataMap[pickle.Uri] = featureData
		}

		// Build scenario data
		scenarioData := ScenarioData{
			Keyword:  "Scenario",
			Name:     pickle.Name,
			Status:   "passed", // default
			Duration: "",
			Steps:    []StepData{},
		}

		// Build step data
		hasFailed := false
		for _, step := range pickle.Steps {
			status := "passed"
			if f.stepStatus[pickle.Id] != nil {
				if s, ok := f.stepStatus[pickle.Id][step.Id]; ok {
					status = s
				}
			}

			if status == "failed" {
				hasFailed = true
			}

			// Get step duration
			duration := ""
			if f.stepTiming[pickle.Id] != nil {
				if d, ok := f.stepTiming[pickle.Id][step.Id]; ok {
					duration = d.String()
				}
			}

			scenarioData.Steps = append(scenarioData.Steps, StepData{
				Keyword:      step.AstNodeIds[0], // This is a simplification
				Name:         step.Text,
				Status:       status,
				Duration:     duration,
				ErrorMessage: "",
			})
		}

		// Update scenario status based on steps
		if hasFailed {
			scenarioData.Status = "failed"
			data.FailedScenarios++
		} else {
			data.PassedScenarios++
		}

		featureData.Scenarios = append(featureData.Scenarios, scenarioData)
	}

	// Convert map to slice
	for _, fd := range featureDataMap {
		data.Features = append(data.Features, *fd)
	}

	t := template.Must(template.New("report").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error generating HTML: %v", err)
	}

	return buf.String()
}

type FeatureData struct {
	Name      string
	Scenarios []ScenarioData
}

type ScenarioData struct {
	Keyword  string
	Name     string
	Status   string
	Duration string
	Steps    []StepData
}

type StepData struct {
	Keyword      string
	Name         string
	Status       string
	Duration     string
	ErrorMessage string
}

// FormatterFunc creates a new HTML formatter
func FormatterFunc(suite string, out io.Writer) formatters.Formatter {
	return &HTMLFormatter{
		out:        out,
		stepStatus: make(map[string]map[string]string),
		stepTiming: make(map[string]map[string]time.Duration),
		stepStart:  make(map[string]map[string]time.Time),
	}
}

// WriteHTMLReport is a helper to write HTML to a file from JSON
func WriteHTMLReport(jsonPath, htmlPath string) error {
	// Read JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON: %w", err)
	}

	// Parse JSON (using existing struct from reporter.go)
	var report interface{}
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// For now, just write a simple HTML
	// TODO: Convert JSON to proper HTML using the template
	htmlContent := "<!DOCTYPE html><html><body><h1>Report</h1><pre>" + string(data) + "</pre></body></html>"

	return os.WriteFile(htmlPath, []byte(htmlContent), 0644)
}
