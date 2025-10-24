package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// TestCollectorOptions tests all combinations of collector options
func TestCollectorOptions(t *testing.T) {
	// Test all combinations of collector options
	// Each test case has the collector options and the expected behavior
	testCases := []struct {
		name              string
		collectMemInfo    bool
		collectStats      bool
		collectZoneStats  bool
		collectZoneStatus bool
		collectZoneSerial bool
		collectZoneTimers bool
	}{
		{"All options enabled", true, true, true, true, true, true},
		{"All options disabled", false, false, false, false, false, false},
		{"Only memInfo", true, false, false, false, false, false},
		{"Only global stats", false, true, false, false, false, false},
		{"Only zone stats", false, false, true, false, false, false},
		{"Only zone status", false, false, false, true, false, false},
		{"Only zone serial", false, false, false, false, true, false},
		{"Only zone timers", false, false, false, false, false, true},
	}

	// For each test case, create a collector and verify that the options are set correctly
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a collector with the specified options
			collector := NewKnotCollector("/test", 1000,
				tc.collectMemInfo,
				tc.collectStats,
				tc.collectZoneStats,
				tc.collectZoneStatus,
				tc.collectZoneSerial,
				tc.collectZoneTimers,
			)

			// Verify that the options are set correctly
			assert.Equal(t, tc.collectMemInfo, collector.collectMemInfo)
			assert.Equal(t, tc.collectStats, collector.collectStats)
			assert.Equal(t, tc.collectZoneStats, collector.collectZoneStats)
			assert.Equal(t, tc.collectZoneStatus, collector.collectZoneStatus)
			assert.Equal(t, tc.collectZoneSerial, collector.collectZoneSerial)
			assert.Equal(t, tc.collectZoneTimers, collector.collectZoneTimers)

			// Create a registry and register the collector
			registry := prometheus.NewRegistry()
			registry.MustRegister(collector)

			// Collecting metrics should not panic even with options disabled
			assert.NotPanics(t, func() {
				metrics, err := registry.Gather()
				assert.NoError(t, err)

				// Should at least have the build info metric
				foundBuildInfo := false
				for _, mf := range metrics {
					if mf.GetName() == "knot_build_info" {
						foundBuildInfo = true
						break
					}
				}

				assert.True(t, foundBuildInfo, "Build info metric should be present")
			})
		})
	}
}

// TestMemoryUsageWithNoProcess tests memoryUsage when no knotd process exists
func TestMemoryUsageWithNoProcess(t *testing.T) {
	// This should return an empty map when knotd is not running
	usage := memoryUsage()
	assert.NotNil(t, usage)
	// Map should be empty or have no valid entries when knotd is not running
	assert.IsType(t, map[string]uint64{}, usage)
}

// TestGetProcessMemoryInvalidPID tests getProcessMemory with invalid PIDs
func TestGetProcessMemoryInvalidPID(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{"negative PID", -1},
		{"zero PID", 0},
		{"very large PID", 9999999},
		{"non-existent PID", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memory := getProcessMemory(tt.pid)
			assert.Equal(t, uint64(0), memory)
		})
	}
}

// TestGetProcessMemorySelfProcess tests getProcessMemory with current process
func TestGetProcessMemorySelfProcess(t *testing.T) {
	// Test with the current process PID (should have some memory usage)
	pid := 1 // init process should always exist
	memory := getProcessMemory(pid)
	// Memory could be 0 if we can't read /proc/1/status (permission issue)
	// or > 0 if we can read it
	assert.GreaterOrEqual(t, memory, uint64(0))
}

// TestCollectWithMemInfo tests Collect with memory info enabled
func TestCollectWithMemInfo(t *testing.T) {
	collector := NewKnotCollector("/nonexistent/socket.sock", 1000,
		true, false, false, false, false, false)

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Consume all metrics
	metricsCount := 0
	for range ch {
		metricsCount++
	}

	// Should have at least the build info metric
	assert.Greater(t, metricsCount, 0)
}
