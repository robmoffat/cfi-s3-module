package reporters

import (
	"io"

	"github.com/cucumber/godog/formatters"
)

// PortTestParams holds the parameters for port testing
type PortTestParams struct {
	PortNumber  string
	HostName    string
	Protocol    string
	ServiceType string
}

// FormatterFactory creates formatters with embedded test parameters
type FormatterFactory struct {
	params PortTestParams
}

// NewFormatterFactory creates a new formatter factory with the given parameters
func NewFormatterFactory(params PortTestParams) *FormatterFactory {
	return &FormatterFactory{
		params: params,
	}
}

// GetHTMLFormatterFunc returns a configured HTML formatter function
func (ff *FormatterFactory) GetHTMLFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewHTMLFormatterWithParams(suite, out, ff.params)
	}
}

// GetOCSFFormatterFunc returns a configured OCSF formatter function
func (ff *FormatterFactory) GetOCSFFormatterFunc() func(string, io.Writer) formatters.Formatter {
	return func(suite string, out io.Writer) formatters.Formatter {
		return NewOCSFFormatterWithParams(suite, out, ff.params)
	}
}
