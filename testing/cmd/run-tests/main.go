package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/finos-labs/ccc-cfi-compliance/testing/inspection"
	"github.com/finos-labs/ccc-cfi-compliance/testing/language/cloud"
)

func main() {
	// Define command-line flags
	provider := flag.String("provider", "", "Cloud provider (aws, azure, or gcp)")
	outputDir := flag.String("output", "output", "Output directory for test reports")
	featuresPath := flag.String("features", "testing/features", "Path to feature files")
	timeout := flag.Duration("timeout", 30*time.Minute, "Timeout for all tests")
	skipPorts := flag.Bool("skip-ports", false, "Skip port tests")
	skipServices := flag.Bool("skip-services", false, "Skip service tests")

	flag.Parse()

	// Validate required flags
	if *provider == "" {
		log.Fatal("Error: -provider flag is required (aws, azure, or gcp)")
	}

	if *provider != "aws" && *provider != "azure" && *provider != "gcp" {
		log.Fatalf("Error: invalid provider '%s' (must be aws, azure, or gcp)", *provider)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	log.Printf("🚀 Starting CCC CFI Compliance Tests")
	log.Printf("   Provider: %s", *provider)
	log.Printf("   Output Directory: %s", *outputDir)
	log.Printf("   Features Path: %s", *featuresPath)
	log.Printf("   Timeout: %s", *timeout)
	log.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	totalTests := 0
	passedTests := 0
	failedTests := 0
	skippedTests := 0

	// Run port tests
	if !*skipPorts {
		log.Println("🔍 Discovering accessible ports...")
		ports, err := inspection.GetAccessiblePorts(ctx, *provider)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to discover ports: %v", err)
		} else {
			log.Printf("   Found %d accessible port(s)", len(ports))

			for i, port := range ports {
				log.Printf("\n📡 Running tests for port %d/%d:", i+1, len(ports))
				log.Printf("   Host: %s", port.HostName)
				log.Printf("   Port: %s", port.PortNumber)
				log.Printf("   Protocol: %s", port.Protocol)
				log.Printf("   Provider Service: %s", port.ProviderServiceType)
				log.Printf("   Catalog Type: %s", port.CatalogType)

				totalTests++
				result := runPortTest(port, *featuresPath, *outputDir)

				switch result {
				case "passed":
					passedTests++
					log.Printf("   ✅ PASSED")
				case "failed":
					failedTests++
					log.Printf("   ❌ FAILED")
				case "skipped":
					skippedTests++
					log.Printf("   ⏭️  SKIPPED")
				}
			}
		}
	} else {
		log.Println("⏭️  Skipping port tests")
	}

	// Run service tests
	if !*skipServices {
		log.Println("\n🔍 Discovering services...")
		services, err := inspection.GetServices(ctx, *provider)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to discover services: %v", err)
		} else {
			log.Printf("   Found %d service(s)", len(services))

			for i, service := range services {
				// Skip services without a catalog type
				if service.CatalogType == "" {
					log.Printf("\n⏭️  Skipping service %d/%d (no catalog type mapping):", i+1, len(services))
					log.Printf("   Provider Service: %s", service.ProviderServiceType)
					continue
				}

				log.Printf("\n🛠️  Running tests for service %d/%d:", i+1, len(services))
				log.Printf("   Host: %s", service.HostName)
				log.Printf("   Provider Service: %s", service.ProviderServiceType)
				log.Printf("   Catalog Type: %s", service.CatalogType)
				log.Printf("   Region: %s", service.Region)
				log.Printf("   UID: %s", service.UID)

				totalTests++
				result := runServiceTest(service, *featuresPath, *outputDir)

				switch result {
				case "passed":
					passedTests++
					log.Printf("   ✅ PASSED")
				case "failed":
					failedTests++
					log.Printf("   ❌ FAILED")
				case "skipped":
					skippedTests++
					log.Printf("   ⏭️  SKIPPED")
				}
			}
		}
	} else {
		log.Println("⏭️  Skipping service tests")
	}

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("📊 Test Summary")
	log.Printf("   Total Tests: %d", totalTests)
	log.Printf("   Passed: %d", passedTests)
	log.Printf("   Failed: %d", failedTests)
	log.Printf("   Skipped: %d", skippedTests)
	log.Println(strings.Repeat("=", 60))

	// Exit with appropriate code
	if failedTests > 0 {
		log.Println("❌ Some tests failed")
		os.Exit(1)
	} else if totalTests == 0 {
		log.Println("⚠️  No tests were run")
		os.Exit(1)
	} else {
		log.Println("✅ All tests passed")
		os.Exit(0)
	}
}

// runPortTest runs tests for a single port configuration
func runPortTest(port inspection.TestParams, featuresPath, outputDir string) string {
	// Create a safe filename from the port details
	filename := fmt.Sprintf("port_%s_%s_%s_%s",
		sanitizeFilename(port.Provider),
		sanitizeFilename(port.HostName),
		sanitizeFilename(port.PortNumber),
		sanitizeFilename(port.Protocol),
	)
	reportPath := filepath.Join(outputDir, filename)

	// Create a mock testing.T for the test runner
	mockT := &mockTestingT{
		name: filename,
	}

	// Run the port tests with panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("   ⚠️  Test panicked: %v", r)
			mockT.failed = true
		}
	}()

	cloud.RunPortTests(&mockT.T, port, featuresPath, reportPath)

	// Determine result based on mock testing state
	if mockT.skipped {
		return "skipped"
	} else if mockT.failed {
		return "failed"
	}
	return "passed"
}

// runServiceTest runs tests for a single service configuration
func runServiceTest(service inspection.TestParams, featuresPath, outputDir string) string {
	// Create a safe filename from the service details
	filename := fmt.Sprintf("service_%s_%s_%s",
		sanitizeFilename(service.Provider),
		sanitizeFilename(service.CatalogType),
		sanitizeFilename(service.UID),
	)
	reportPath := filepath.Join(outputDir, filename)

	// Create a mock testing.T for the test runner
	mockT := &mockTestingT{
		name: filename,
	}

	// Run the service tests with panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("   ⚠️  Test panicked: %v", r)
			mockT.failed = true
		}
	}()

	cloud.RunServiceTests(&mockT.T, service, featuresPath, reportPath)

	// Determine result based on mock testing state
	if mockT.skipped {
		return "skipped"
	} else if mockT.failed {
		return "failed"
	}
	return "passed"
}

// sanitizeFilename removes characters that aren't safe for filenames
func sanitizeFilename(s string) string {
	// Replace common unsafe characters
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "*", "_")
	s = strings.ReplaceAll(s, "?", "_")
	s = strings.ReplaceAll(s, "\"", "_")
	s = strings.ReplaceAll(s, "<", "_")
	s = strings.ReplaceAll(s, ">", "_")
	s = strings.ReplaceAll(s, "|", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")

	// Truncate if too long
	if len(s) > 100 {
		s = s[:100]
	}

	return s
}

// mockTestingT wraps a real testing.T to capture results when run outside go test
type mockTestingT struct {
	testing.T
	failed  bool
	skipped bool
	name    string
}

// Override methods to capture state without calling the embedded T
func (m *mockTestingT) Error(args ...interface{}) { m.failed = true; log.Print(args...) }
func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.failed = true
	log.Printf(format, args...)
}
func (m *mockTestingT) Fail()        { m.failed = true }
func (m *mockTestingT) FailNow()     { m.failed = true; panic("FailNow") }
func (m *mockTestingT) Failed() bool { return m.failed }
func (m *mockTestingT) Fatal(args ...interface{}) {
	m.failed = true
	log.Print(args...)
	panic("Fatal")
}
func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	log.Printf(format, args...)
	panic("Fatalf")
}
func (m *mockTestingT) Log(args ...interface{})                 { log.Print(args...) }
func (m *mockTestingT) Logf(format string, args ...interface{}) { log.Printf(format, args...) }
func (m *mockTestingT) Name() string                            { return m.name }
func (m *mockTestingT) Skip(args ...interface{})                { m.skipped = true; log.Print(args...) }
func (m *mockTestingT) SkipNow()                                { m.skipped = true; panic("SkipNow") }
func (m *mockTestingT) Skipf(format string, args ...interface{}) {
	m.skipped = true
	log.Printf(format, args...)
	panic("Skipf")
}
func (m *mockTestingT) Skipped() bool { return m.skipped }
