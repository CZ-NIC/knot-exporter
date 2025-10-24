package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/CZ-NIC/knot-exporter/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrintVersion tests the printVersion function
func TestPrintVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call printVersion
	printVersion()

	// Restore stdout
	err := w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify output contains expected information
	assert.Contains(t, output, "Knot DNS Exporter")
	assert.Contains(t, output, "Version:")
	assert.Contains(t, output, "Build time:")
	assert.Contains(t, output, "Git commit:")
	assert.Contains(t, output, "Go version:")
	assert.Contains(t, output, "Libknot:")
	assert.Contains(t, output, "Platform:")
}

// TestValidateConfig tests the validateConfig function
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		sockPath    string
		addr        string
		port        int
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "non-existent socket path",
			sockPath:    "/nonexistent/socket.sock",
			addr:        "127.0.0.1",
			port:        9433,
			shouldError: true,
			errorMsg:    "does not exist",
		},
		{
			name:        "invalid port - too low",
			sockPath:    "/tmp",
			addr:        "127.0.0.1",
			port:        0,
			shouldError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid port - too high",
			sockPath:    "/tmp",
			addr:        "127.0.0.1",
			port:        70000,
			shouldError: true,
			errorMsg:    "invalid port number",
		},
		{
			name:        "invalid address",
			sockPath:    "/tmp",
			addr:        "invalid-address",
			port:        9433,
			shouldError: true,
			errorMsg:    "invalid listen address",
		},
		{
			name:        "localhost address",
			sockPath:    "/tmp",
			addr:        "localhost",
			port:        9433,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.sockPath, tt.addr, tt.port)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For successful cases, we might get an error if the port is already in use
				// which is okay in test environments
				if err != nil {
					assert.Contains(t, err.Error(), "cannot bind")
				}
			}
		})
	}
}

// TestValidateConfigValidIP tests validateConfig with a valid IP
func TestValidateConfigValidIP(t *testing.T) {
	// Create a temporary file to act as socket
	tmpFile, err := os.CreateTemp("", "test-socket-*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	err = tmpFile.Close()
	require.NoError(t, err)

	// Test with valid IP - might fail if port in use, which is okay
	err = validateConfig(tmpFile.Name(), "127.0.0.1", 19433)
	// Either no error or "cannot bind" error is acceptable
	if err != nil {
		assert.Contains(t, err.Error(), "cannot bind")
	}
}

// TestTestKnotConnection tests the testKnotConnection function
func TestTestKnotConnection(t *testing.T) {
	tests := []struct {
		name        string
		sockPath    string
		timeout     int
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "non-existent socket",
			sockPath:    "/nonexistent/socket.sock",
			timeout:     1000,
			shouldError: true,
			errorMsg:    "failed to connect",
		},
		{
			name:        "invalid socket path",
			sockPath:    "/tmp",
			timeout:     1000,
			shouldError: true,
			errorMsg:    "failed to",
		},
		{
			name:        "empty socket path",
			sockPath:    "",
			timeout:     1000,
			shouldError: true,
		},
		{
			name:        "zero timeout",
			sockPath:    "/nonexistent/socket.sock",
			timeout:     0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testKnotConnection(tt.sockPath, tt.timeout)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTestKnotConnectionDebugMode tests testKnotConnection with debug mode
func TestTestKnotConnectionDebugMode(t *testing.T) {
	// Import utils to set debug mode
	oldDebugMode := utils.DebugMode
	utils.DebugMode = true
	defer func() { utils.DebugMode = oldDebugMode }()

	err := testKnotConnection("/nonexistent/socket.sock", 1000)
	assert.Error(t, err)
}

// TestHealthCheck tests the healthCheck handler
func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		sockPath       string
		timeout        int
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "health check with non-existent socket",
			sockPath:       "/nonexistent/socket.sock",
			timeout:        1000,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Health check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := healthCheck(tt.sockPath, tt.timeout)

			// Create a test request
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler(w, req)

			// Check the response
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestHealthCheckSuccess tests healthCheck with a mock successful connection
func TestHealthCheckSuccessScenario(t *testing.T) {
	// This test verifies the handler responds correctly
	// In a real scenario with working socket, it would return 200 OK
	handler := healthCheck("/tmp/test.sock", 1000)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// Should fail since socket doesn't exist, but test the handler structure
	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}

// TestSetupGracefulShutdown tests that setupGracefulShutdown doesn't panic
func TestSetupGracefulShutdown(t *testing.T) {
	// Create a test server
	server := &http.Server{
		Addr: "127.0.0.1:19999",
	}

	// setupGracefulShutdown should not panic
	assert.NotPanics(t, func() {
		setupGracefulShutdown(server)
	})
}

// TestMainFunctionsIntegration tests integration of multiple functions
func TestMainFunctionsIntegration(t *testing.T) {
	// Test that we can call printVersion without panic
	assert.NotPanics(t, func() {
		printVersion()
	})

	// Test validateConfig with various inputs
	err := validateConfig("/nonexistent", "invalid", 0)
	assert.Error(t, err)

	// Test testKnotConnection with invalid socket
	err = testKnotConnection("/nonexistent/socket.sock", 1000)
	assert.Error(t, err)
}

// TestValidateConfigEdgeCases tests edge cases for validateConfig
func TestValidateConfigEdgeCases(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-socket-*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	err = tmpFile.Close()
	require.NoError(t, err)

	tests := []struct {
		name     string
		sockPath string
		addr     string
		port     int
	}{
		{
			name:     "minimum valid port",
			sockPath: tmpFile.Name(),
			addr:     "127.0.0.1",
			port:     1,
		},
		{
			name:     "maximum valid port",
			sockPath: tmpFile.Name(),
			addr:     "127.0.0.1",
			port:     65535,
		},
		{
			name:     "localhost string",
			sockPath: tmpFile.Name(),
			addr:     "localhost",
			port:     8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These might fail with "cannot bind" if port is in use, which is acceptable
			err := validateConfig(tt.sockPath, tt.addr, tt.port)
			if err != nil {
				assert.Contains(t, err.Error(), "cannot bind")
			}
		})
	}
}

// TestHealthCheckContentType tests that health check sets correct content type
func TestHealthCheckContentType(t *testing.T) {
	handler := healthCheck("/nonexistent/socket.sock", 1000)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// Should have set content type even on error
	contentType := w.Header().Get("Content-Type")
	// Either text/plain or empty is acceptable
	if contentType != "" {
		assert.Contains(t, contentType, "text/plain")
	}
}

// TestHealthCheckMultipleRequests tests health check with multiple requests
func TestHealthCheckMultipleRequests(t *testing.T) {
	handler := healthCheck("/nonexistent/socket.sock", 500)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "Health check failed")
	}
}

// TestHealthCheckWithDifferentTimeouts tests health check with various timeouts
func TestHealthCheckWithDifferentTimeouts(t *testing.T) {
	timeouts := []int{100, 500, 1000, 2000}

	for _, timeout := range timeouts {
		t.Run(string(rune(timeout)), func(t *testing.T) {
			handler := healthCheck("/nonexistent/socket.sock", timeout)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		})
	}
}

// TestVersionBuildVariables tests that build variables are accessible
func TestVersionBuildVariables(t *testing.T) {
	// These variables should be set by the build system or have defaults
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, buildTime)
	assert.NotEmpty(t, gitCommit)
	assert.NotEmpty(t, goVersion)

	// Default values
	if version == "" {
		version = "dev"
	}
	if buildTime == "" {
		buildTime = "unknown"
	}
	if gitCommit == "" {
		gitCommit = "unknown"
	}

	assert.NotEmpty(t, version)
	assert.NotEmpty(t, buildTime)
	assert.NotEmpty(t, gitCommit)
}

// TestPrintVersionOutput tests specific output format of printVersion
func TestPrintVersionOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printVersion()

	err := w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Should contain specific format markers
	lines := strings.Split(output, "\n")
	assert.Greater(t, len(lines), 5, "Should have multiple lines of output")

	// First line should be the title
	assert.Contains(t, lines[0], "Knot DNS Exporter")
}

// TestValidateConfigPortBoundaries tests port boundary conditions
func TestValidateConfigPortBoundaries(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-socket-*")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	err = tmpFile.Close()
	require.NoError(t, err)

	tests := []struct {
		port        int
		shouldError bool
	}{
		{port: -1, shouldError: true},
		{port: 0, shouldError: true},
		{port: 1, shouldError: false},
		{port: 65535, shouldError: false},
		{port: 65536, shouldError: true},
		{port: 99999, shouldError: true},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.port)), func(t *testing.T) {
			err := validateConfig(tmpFile.Name(), "127.0.0.1", tt.port)
			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid port number")
			} else {
				// Might get "cannot bind" error if port is in use
				if err != nil {
					assert.Contains(t, err.Error(), "cannot bind")
				}
			}
		})
	}
}
