package cloud

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/cucumber/godog"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/generic"
)

// CloudWorld extends PropsWorld with cloud-specific functionality
type CloudWorld struct {
	*generic.PropsWorld
	mu sync.RWMutex
}

// NewCloudWorld creates a new CloudWorld instance
func NewCloudWorld() *CloudWorld {
	return &CloudWorld{
		PropsWorld: generic.NewPropsWorld(),
	}
}

// RegisterSteps registers all cloud-specific step definitions
func (cw *CloudWorld) RegisterSteps(ctx *godog.ScenarioContext) {
	// Register generic steps first
	cw.PropsWorld.RegisterSteps(ctx)

	// Cloud-specific steps matching Cucumber-Cloud-Language.md

	// OpenSSL connections
	ctx.Step(`^an openssl s_client request to "([^"]*)" on "([^"]*)" protocol "([^"]*)" as "([^"]*)"$`, cw.opensslClientRequestWithProtocol)
	ctx.Step(`^an openssl s_client request using "([^"]*)" to "([^"]*)" on "([^"]*)" protocol "([^"]*)" as "([^"]*)"$`, cw.opensslClientRequestWithTLSAndProtocol)

	// Plain client connections
	ctx.Step(`^a client connects to "([^"]*)" with protocol "([^"]*)" on port "([^"]*)" as "([^"]*)"$`, cw.clientConnectsWithProtocol)

	// Connection operations
	ctx.Step(`^I transmit "([^"]*)" over "([^"]*)"$`, cw.transmitOverConnection)
	ctx.Step(`^close connection "([^"]*)"$`, cw.closeConnection)
	ctx.Step(`^"([^"]*)" is closed$`, cw.connectionIsClosed)

	// SSL Support reports
	ctx.Step(`^"([^"]*)" contains details of SSL Support type "([^"]*)" for "([^"]*)" on port "([^"]*)"$`, cw.getSSLSupportReport)
	ctx.Step(`^"([^"]*)" contains details of SSL Support type "([^"]*)" for "([^"]*)" on port "([^"]*)" with STARTTLS$`, cw.getSSLSupportReportWithSTARTTLS)
}

// opensslClientRequest creates an OpenSSL s_client connection with optional TLS version
func (cw *CloudWorld) opensslClientRequest(tlsVersion, port, hostName, protocol, connectionName string) error {
	tlsVersionResolved := cw.HandleResolve(tlsVersion)
	portResolved := cw.HandleResolve(port)
	hostResolved := cw.HandleResolve(hostName)
	connectionNameResolved := cw.HandleResolve(connectionName)

	// Build openssl s_client command
	args := []string{"s_client", "-connect", fmt.Sprintf("%v:%v", hostResolved, portResolved)}

	// Add TLS version if specified
	if tlsVersionResolved != nil && fmt.Sprintf("%v", tlsVersionResolved) != "" {
		args = append(args, "-"+fmt.Sprintf("%v", tlsVersionResolved))
	}

	cmd := exec.Command("openssl", args...)
	output, err := cmd.CombinedOutput()

	cw.Props[fmt.Sprintf("%v", connectionNameResolved)] = string(output)
	if err != nil {
		cw.Props["result"] = err
	} else {
		cw.Props["result"] = string(output)
	}

	return nil
}

// opensslClientRequestWithProtocol creates an OpenSSL s_client connection
func (cw *CloudWorld) opensslClientRequestWithProtocol(port, hostName, protocol, connectionName string) error {
	return cw.opensslClientRequest("", port, hostName, protocol, connectionName)
}

// opensslClientRequestWithTLSAndProtocol creates an OpenSSL s_client connection with specific TLS version
func (cw *CloudWorld) opensslClientRequestWithTLSAndProtocol(tlsVersion, port, hostName, protocol, connectionName string) error {
	return cw.opensslClientRequest(tlsVersion, port, hostName, protocol, connectionName)
}

// clientConnectsWithProtocol establishes a plain client connection to a host with a specific protocol
func (cw *CloudWorld) clientConnectsWithProtocol(hostName, protocol, port, connectionName string) error {
	return cw.opensslClientRequest("", port, hostName, "", connectionName)
}

// transmitOverConnection sends data over an established connection
func (cw *CloudWorld) transmitOverConnection(data, connectionName string) error {
	dataResolved := cw.HandleResolve(data)
	connectionNameResolved := cw.HandleResolve(connectionName)

	if dataResolved == nil {
		return fmt.Errorf("data %s not found", data)
	}

	// For now, we just store the transmission intent
	// In a real implementation, this would send data over an active connection
	cw.Props["result"] = fmt.Sprintf("Transmitted: %v over %v", dataResolved, connectionNameResolved)
	return nil
}

// closeConnection closes an established connection
func (cw *CloudWorld) closeConnection(connectionName string) error {
	connectionNameResolved := cw.HandleResolve(connectionName)

	connection := cw.Props[fmt.Sprintf("%v", connectionNameResolved)]
	if connection == nil {
		return fmt.Errorf("connection %s not found", connectionName)
	}

	// Mark connection as closed
	delete(cw.Props, fmt.Sprintf("%v", connectionNameResolved))
	return nil
}

// connectionIsClosed verifies that a connection has been closed
func (cw *CloudWorld) connectionIsClosed(connectionName string) error {
	connectionNameResolved := cw.HandleResolve(connectionName)

	connection := cw.Props[fmt.Sprintf("%v", connectionNameResolved)]
	if connection != nil {
		return fmt.Errorf("connection %s is still open", connectionName)
	}
	return nil
}

// runTestSSL is a helper function to run testssl.sh and return JSON report
func (cw *CloudWorld) runTestSSL(reportName, testType, hostName, port string, useSTARTTLS bool) error {
	reportNameResolved := cw.HandleResolve(reportName)
	testTypeResolved := cw.HandleResolve(testType)
	hostResolved := cw.HandleResolve(hostName)
	portResolved := cw.HandleResolve(port)

	// Create temporary file for JSON output
	tempFile := fmt.Sprintf("/tmp/testssl_%v_%v_%v", testTypeResolved, hostResolved, portResolved)
	if useSTARTTLS {
		tempFile += "_starttls"
	}
	tempFile += ".json"

	// Build testssl.sh command
	testsslPath := "./testssl.sh"
	args := []string{testsslPath, "--" + fmt.Sprintf("%v", testTypeResolved)}

	if useSTARTTLS {
		// Determine STARTTLS protocol from port
		protocol := cw.HandleResolve("{protocol}")
		if protocol == nil {
			protocol = "smtp" // default
		}
		args = append(args, "-t", fmt.Sprintf("%v", protocol))
	}

	args = append(args, "--jsonfile", tempFile, fmt.Sprintf("%v:%v", hostResolved, portResolved))
	cmd := exec.Command("bash", args...)

	_, err := cmd.CombinedOutput()
	if err != nil {
		// testssl.sh might return non-zero exit code even on success
		// Continue to try reading the JSON file
	}

	// Read and parse JSON output
	jsonData, err := exec.Command("cat", tempFile).Output()
	if err != nil {
		return fmt.Errorf("failed to read testssl.sh output: %v", err)
	}

	var report interface{}
	if err := json.Unmarshal(jsonData, &report); err != nil {
		return fmt.Errorf("failed to parse testssl.sh JSON: %v", err)
	}

	cw.Props[fmt.Sprintf("%v", reportNameResolved)] = report
	return nil
}

// getSSLSupportReport runs testssl.sh and returns JSON report
func (cw *CloudWorld) getSSLSupportReport(reportName, testType, hostName, port string) error {
	return cw.runTestSSL(reportName, testType, hostName, port, false)
}

// getSSLSupportReportWithSTARTTLS runs testssl.sh with STARTTLS support
func (cw *CloudWorld) getSSLSupportReportWithSTARTTLS(reportName, testType, hostName, port string) error {
	return cw.runTestSSL(reportName, testType, hostName, port, true)
}
