package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"
)

// CucumberReport represents the structure of Cucumber JSON report
type CucumberReport []Feature

type Feature struct {
	URI         string     `json:"uri"`
	ID          string     `json:"id"`
	Keyword     string     `json:"keyword"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Line        int        `json:"line"`
	Elements    []Scenario `json:"elements"`
}

type Scenario struct {
	ID          string `json:"id"`
	Keyword     string `json:"keyword"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Type        string `json:"type"`
	Tags        []Tag  `json:"tags"`
	Steps       []Step `json:"steps"`
}

type Tag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type Step struct {
	Keyword string     `json:"keyword"`
	Name    string     `json:"name"`
	Line    int        `json:"line"`
	Match   StepMatch  `json:"match"`
	Result  StepResult `json:"result"`
}

type StepMatch struct {
	Location string `json:"location"`
}

type StepResult struct {
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
	Error    string `json:"error_message,omitempty"`
}

// ReportStats holds summary statistics
type ReportStats struct {
	TotalFeatures   int
	TotalScenarios  int
	PassedScenarios int
	FailedScenarios int
	TotalSteps      int
	PassedSteps     int
	FailedSteps     int
	SkippedSteps    int
	TotalDuration   time.Duration
	GeneratedAt     string
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cucumber Test Report</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            padding: 30px;
            background-color: #f8f9fa;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            margin-bottom: 5px;
        }
        .stat-label {
            color: #666;
            font-size: 0.9em;
        }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .content {
            padding: 30px;
        }
        .feature {
            margin-bottom: 40px;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            overflow: hidden;
        }
        .feature-header {
            background-color: #343a40;
            color: white;
            padding: 20px;
        }
        .feature-title {
            font-size: 1.5em;
            margin: 0;
        }
        .feature-description {
            margin: 10px 0 0 0;
            opacity: 0.9;
            white-space: pre-line;
        }
        .scenario {
            border-bottom: 1px solid #e9ecef;
        }
        .scenario:last-child {
            border-bottom: none;
        }
        .scenario-header {
            padding: 15px 20px;
            background-color: #f8f9fa;
            border-left: 4px solid #6c757d;
        }
        .scenario-header.passed {
            border-left-color: #28a745;
        }
        .scenario-header.failed {
            border-left-color: #dc3545;
        }
        .scenario-title {
            font-weight: bold;
            margin-bottom: 5px;
        }
        .scenario-tags {
            font-size: 0.8em;
            color: #666;
        }
        .tag {
            background-color: #e9ecef;
            padding: 2px 8px;
            border-radius: 12px;
            margin-right: 5px;
        }
        .steps {
            padding: 0;
        }
        .step {
            padding: 10px 20px;
            border-left: 4px solid transparent;
            display: flex;
            align-items: center;
        }
        .step.passed {
            border-left-color: #28a745;
            background-color: #f8fff9;
        }
        .step.failed {
            border-left-color: #dc3545;
            background-color: #fff8f8;
        }
        .step.skipped {
            border-left-color: #ffc107;
            background-color: #fffdf5;
        }
        .step-keyword {
            font-weight: bold;
            margin-right: 8px;
            min-width: 60px;
        }
        .step-name {
            flex: 1;
        }
        .step-duration {
            font-size: 0.8em;
            color: #666;
            margin-left: 10px;
        }
        .step-error {
            margin-top: 10px;
            padding: 10px;
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            border-radius: 4px;
            color: #721c24;
            font-family: monospace;
            font-size: 0.9em;
            white-space: pre-wrap;
        }
        .footer {
            text-align: center;
            padding: 20px;
            background-color: #f8f9fa;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Cucumber Test Report</h1>
            <p>Generated on {{.GeneratedAt}}</p>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">{{.TotalFeatures}}</div>
                <div class="stat-label">Features</div>
            </div>
            <div class="stat-card">
                <div class="stat-number passed">{{.PassedScenarios}}</div>
                <div class="stat-label">Passed Scenarios</div>
            </div>
            <div class="stat-card">
                <div class="stat-number failed">{{.FailedScenarios}}</div>
                <div class="stat-label">Failed Scenarios</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.TotalSteps}}</div>
                <div class="stat-label">Total Steps</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.TotalDuration}}</div>
                <div class="stat-label">Duration</div>
            </div>
        </div>

        <div class="content">
            {{range .Features}}
            <div class="feature">
                <div class="feature-header">
                    <h2 class="feature-title">{{.Name}}</h2>
                    {{if .Description}}
                    <p class="feature-description">{{.Description}}</p>
                    {{end}}
                </div>
                
                {{range .Elements}}
                <div class="scenario">
                    <div class="scenario-header {{.Status}}">
                        <div class="scenario-title">{{.Keyword}}: {{.Name}}</div>
                        {{if .Tags}}
                        <div class="scenario-tags">
                            {{range .Tags}}<span class="tag">{{.Name}}</span>{{end}}
                        </div>
                        {{end}}
                    </div>
                    
                    <div class="steps">
                        {{range .Steps}}
                        <div class="step {{.Result.Status}}">
                            <span class="step-keyword">{{.Keyword}}</span>
                            <span class="step-name">{{.Name}}</span>
                            <span class="step-duration">{{.FormattedDuration}}</span>
                        </div>
                        {{if .Result.Error}}
                        <div class="step-error">{{.Result.Error}}</div>
                        {{end}}
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <div class="footer">
            Report generated by Custom Cucumber HTML Reporter
        </div>
    </div>
</body>
</html>
`

// EnhancedScenario adds computed fields for template rendering
type EnhancedScenario struct {
	Scenario
	Status string
	Steps  []EnhancedStep
}

// EnhancedStep adds computed fields for template rendering
type EnhancedStep struct {
	Step
	FormattedDuration string
}

// EnhancedFeature adds computed fields for template rendering
type EnhancedFeature struct {
	Feature
	Elements []EnhancedScenario
}

// TemplateData holds all data for the HTML template
type TemplateData struct {
	ReportStats
	Features []EnhancedFeature
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run reporter.go <cucumber-json-file> [output-html-file]")
		os.Exit(1)
	}

	jsonFile := os.Args[1]
	htmlFile := "report.html"
	if len(os.Args) > 2 {
		htmlFile = os.Args[2]
	}

	// Read JSON file
	jsonData, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var report CucumberReport
	if err := json.Unmarshal(jsonData, &report); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate HTML report
	if err := generateHTMLReport(report, htmlFile); err != nil {
		fmt.Printf("Error generating HTML report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("HTML report generated: %s\n", htmlFile)
}

func generateHTMLReport(report CucumberReport, outputFile string) error {
	// Calculate statistics
	stats := calculateStats(report)

	// Enhance data for template
	enhancedFeatures := make([]EnhancedFeature, len(report))
	for i, feature := range report {
		enhancedScenarios := make([]EnhancedScenario, len(feature.Elements))
		for j, scenario := range feature.Elements {
			enhancedSteps := make([]EnhancedStep, len(scenario.Steps))
			scenarioStatus := "passed"

			for k, step := range scenario.Steps {
				enhancedSteps[k] = EnhancedStep{
					Step:              step,
					FormattedDuration: formatDuration(time.Duration(step.Result.Duration) * time.Nanosecond),
				}

				if step.Result.Status == "failed" {
					scenarioStatus = "failed"
				}
			}

			enhancedScenarios[j] = EnhancedScenario{
				Scenario: scenario,
				Status:   scenarioStatus,
				Steps:    enhancedSteps,
			}
		}

		enhancedFeatures[i] = EnhancedFeature{
			Feature:  feature,
			Elements: enhancedScenarios,
		}
	}

	templateData := TemplateData{
		ReportStats: stats,
		Features:    enhancedFeatures,
	}

	// Parse and execute template
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

func calculateStats(report CucumberReport) ReportStats {
	stats := ReportStats{
		TotalFeatures: len(report),
		GeneratedAt:   time.Now().Format("January 2, 2006 at 3:04 PM"),
	}

	var totalDuration time.Duration

	for _, feature := range report {
		for _, scenario := range feature.Elements {
			stats.TotalScenarios++
			scenarioFailed := false

			for _, step := range scenario.Steps {
				stats.TotalSteps++
				totalDuration += time.Duration(step.Result.Duration) * time.Nanosecond

				switch step.Result.Status {
				case "passed":
					stats.PassedSteps++
				case "failed":
					stats.FailedSteps++
					scenarioFailed = true
				case "skipped":
					stats.SkippedSteps++
				}
			}

			if scenarioFailed {
				stats.FailedScenarios++
			} else {
				stats.PassedScenarios++
			}
		}
	}

	stats.TotalDuration = totalDuration
	return stats
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}
