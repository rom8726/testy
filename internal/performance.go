package internal

import (
	"fmt"
	"time"
)

// PerformanceSpec defines performance requirements for a request
type PerformanceSpec struct {
	MaxDuration   string `yaml:"maxDuration,omitempty"`   // Maximum allowed duration (e.g., "200ms", "1s")
	MaxMemoryMB   int    `yaml:"maxMemory,omitempty"`     // Maximum memory usage in MB
	MinThroughput int    `yaml:"minThroughput,omitempty"` // Minimum requests per second
	WarnDuration  string `yaml:"warnDuration,omitempty"`  // Warning threshold for duration
	FailOnWarning bool   `yaml:"failOnWarning,omitempty"` // Fail test if a warning threshold exceeded
}

// ParsedPerformanceSpec is the parsed version of PerformanceSpec
type ParsedPerformanceSpec struct {
	MaxDuration   time.Duration
	WarnDuration  time.Duration
	MaxMemoryMB   int
	MinThroughput int
	FailOnWarning bool
}

// ParsePerformanceSpec parses a performance spec
func ParsePerformanceSpec(spec PerformanceSpec) (*ParsedPerformanceSpec, error) {
	parsed := &ParsedPerformanceSpec{
		MaxMemoryMB:   spec.MaxMemoryMB,
		MinThroughput: spec.MinThroughput,
		FailOnWarning: spec.FailOnWarning,
	}

	// Parse max duration
	if spec.MaxDuration != "" {
		duration, err := time.ParseDuration(spec.MaxDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid maxDuration: %w", err)
		}
		parsed.MaxDuration = duration
	}

	// Parse warn duration
	if spec.WarnDuration != "" {
		duration, err := time.ParseDuration(spec.WarnDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid warnDuration: %w", err)
		}
		parsed.WarnDuration = duration
	}

	return parsed, nil
}

// PerformanceMetrics holds performance measurements
type PerformanceMetrics struct {
	Duration      time.Duration
	MemoryUsageMB int
	StatusCode    int
}

// PerformanceResult contains the result of performance validation
type PerformanceResult struct {
	Passed   bool
	Warnings []string
	Errors   []string
	Metrics  PerformanceMetrics
}

// ValidatePerformance validates performance metrics against spec
func ValidatePerformance(metrics PerformanceMetrics, spec *ParsedPerformanceSpec) PerformanceResult {
	result := PerformanceResult{
		Passed:  true,
		Metrics: metrics,
	}

	if spec == nil {
		return result
	}

	// Check max duration
	if spec.MaxDuration > 0 && metrics.Duration > spec.MaxDuration {
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("request duration %s exceeded maximum %s",
				metrics.Duration, spec.MaxDuration))
	}

	// Check warn duration
	if spec.WarnDuration > 0 && metrics.Duration > spec.WarnDuration {
		warning := fmt.Sprintf("request duration %s exceeded warning threshold %s",
			metrics.Duration, spec.WarnDuration)
		result.Warnings = append(result.Warnings, warning)

		if spec.FailOnWarning {
			result.Passed = false
			result.Errors = append(result.Errors, warning)
		}
	}

	// Check memory usage
	if spec.MaxMemoryMB > 0 && metrics.MemoryUsageMB > spec.MaxMemoryMB {
		result.Passed = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("memory usage %d MB exceeded maximum %d MB",
				metrics.MemoryUsageMB, spec.MaxMemoryMB))
	}

	return result
}

// FormatPerformanceResult formats a performance result for logging
func FormatPerformanceResult(result PerformanceResult) string {
	msg := fmt.Sprintf("Duration: %s", result.Metrics.Duration)

	if result.Metrics.MemoryUsageMB > 0 {
		msg += fmt.Sprintf(", Memory: %d MB", result.Metrics.MemoryUsageMB)
	}

	if len(result.Warnings) > 0 {
		msg += fmt.Sprintf(" [WARNINGS: %d]", len(result.Warnings))
	}

	if !result.Passed {
		msg += fmt.Sprintf(" [FAILED: %d errors]", len(result.Errors))
	} else {
		msg += " [PASSED]"
	}

	return msg
}

// PerformanceTracker tracks performance across multiple requests
type PerformanceTracker struct {
	measurements []PerformanceMetrics
	startTime    time.Time
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		measurements: make([]PerformanceMetrics, 0),
		startTime:    time.Now(),
	}
}

// Record records a performance measurement
func (t *PerformanceTracker) Record(metrics PerformanceMetrics) {
	t.measurements = append(t.measurements, metrics)
}

// GetStats returns performance statistics
func (t *PerformanceTracker) GetStats() PerformanceStats {
	if len(t.measurements) == 0 {
		return PerformanceStats{}
	}

	stats := PerformanceStats{
		Count:     len(t.measurements),
		TotalTime: time.Since(t.startTime),
	}

	var totalDuration time.Duration
	var minDuration time.Duration
	var maxDuration time.Duration

	for i, m := range t.measurements {
		totalDuration += m.Duration

		if i == 0 || m.Duration < minDuration {
			minDuration = m.Duration
		}

		if m.Duration > maxDuration {
			maxDuration = m.Duration
		}
	}

	stats.AvgDuration = totalDuration / time.Duration(len(t.measurements))
	stats.MinDuration = minDuration
	stats.MaxDuration = maxDuration

	// Calculate throughput (requests per second)
	if stats.TotalTime > 0 {
		stats.Throughput = float64(stats.Count) / stats.TotalTime.Seconds()
	}

	return stats
}

// PerformanceStats contains aggregated performance statistics
type PerformanceStats struct {
	Count       int
	TotalTime   time.Duration
	AvgDuration time.Duration
	MinDuration time.Duration
	MaxDuration time.Duration
	Throughput  float64 // Requests per second
}

// FormatPerformanceStats formats performance stats for logging
func FormatPerformanceStats(stats PerformanceStats) string {
	return fmt.Sprintf(
		"Requests: %d, Total: %s, Avg: %s, Min: %s, Max: %s, Throughput: %.2f req/s",
		stats.Count,
		stats.TotalTime,
		stats.AvgDuration,
		stats.MinDuration,
		stats.MaxDuration,
		stats.Throughput,
	)
}

// PerformanceReport generates a detailed performance report
type PerformanceReport struct {
	TestName     string
	Stats        PerformanceStats
	Measurements []PerformanceMetrics
	Passed       bool
	Issues       []string
}

// GeneratePerformanceReport generates a report from a tracker
func GeneratePerformanceReport(
	testName string,
	tracker *PerformanceTracker,
	spec *ParsedPerformanceSpec,
) PerformanceReport {
	report := PerformanceReport{
		TestName:     testName,
		Stats:        tracker.GetStats(),
		Measurements: tracker.measurements,
		Passed:       true,
	}

	if spec == nil {
		return report
	}

	// Check throughput
	if spec.MinThroughput > 0 && report.Stats.Throughput < float64(spec.MinThroughput) {
		report.Passed = false
		report.Issues = append(report.Issues,
			fmt.Sprintf("throughput %.2f req/s is below minimum %d req/s",
				report.Stats.Throughput, spec.MinThroughput))
	}

	// Check individual measurements
	for i, m := range tracker.measurements {
		result := ValidatePerformance(m, spec)
		if !result.Passed {
			report.Passed = false
			for _, err := range result.Errors {
				report.Issues = append(report.Issues,
					fmt.Sprintf("request #%d: %s", i+1, err))
			}
		}
	}

	return report
}

// FormatPerformanceReport formats a performance report
func FormatPerformanceReport(report PerformanceReport) string {
	result := fmt.Sprintf("Performance Report: %s\n", report.TestName)
	result += fmt.Sprintf("  %s\n", FormatPerformanceStats(report.Stats))

	if report.Passed {
		result += "  Status: PASSED\n"
	} else {
		result += "  Status: FAILED\n"
		result += "  Issues:\n"
		for _, issue := range report.Issues {
			result += fmt.Sprintf("    - %s\n", issue)
		}
	}

	return result
}
