package internal

import (
	"fmt"
	"math"
	"time"
)

// RetryConfig defines retry behavior for a step
type RetryConfig struct {
	Attempts     int    `yaml:"attempts"`               // Maximum number of attempts (including first try)
	Backoff      string `yaml:"backoff,omitempty"`      // Backoff strategy: linear, exponential, constant
	InitialDelay string `yaml:"initialDelay,omitempty"` // Initial delay (e.g., "100ms", "1s")
	MaxDelay     string `yaml:"maxDelay,omitempty"`     // Maximum delay for exponential backoff
	RetryOn      []int  `yaml:"retryOn,omitempty"`      // Retry only on these status codes
	RetryOnError bool   `yaml:"retryOnError,omitempty"` // Retry on any error
}

// BackoffStrategy defines the type of backoff
type BackoffStrategy int

const (
	BackoffConstant BackoffStrategy = iota
	BackoffLinear
	BackoffExponential
)

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		Attempts:     3,
		Backoff:      "exponential",
		InitialDelay: "100ms",
		MaxDelay:     "10s",
		RetryOnError: true,
	}
}

// ParseRetryConfig parses and validates retry configuration
func ParseRetryConfig(config RetryConfig) (*ParsedRetryConfig, error) {
	parsed := &ParsedRetryConfig{
		Attempts:     config.Attempts,
		RetryOn:      config.RetryOn,
		RetryOnError: config.RetryOnError,
	}

	// Parse backoff strategy
	switch config.Backoff {
	case "", "exponential":
		parsed.Backoff = BackoffExponential
	case "linear":
		parsed.Backoff = BackoffLinear
	case "constant":
		parsed.Backoff = BackoffConstant
	default:
		return nil, fmt.Errorf("unknown backoff strategy: %s", config.Backoff)
	}

	// Parse initial delay
	if config.InitialDelay != "" {
		delay, err := time.ParseDuration(config.InitialDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid initialDelay: %w", err)
		}
		parsed.InitialDelay = delay
	} else {
		parsed.InitialDelay = 100 * time.Millisecond
	}

	// Parse max delay
	if config.MaxDelay != "" {
		maxDelay, err := time.ParseDuration(config.MaxDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid maxDelay: %w", err)
		}
		parsed.MaxDelay = maxDelay
	} else {
		parsed.MaxDelay = 10 * time.Second
	}

	// Validate attempts
	if parsed.Attempts < 1 {
		return nil, fmt.Errorf("attempts must be at least 1")
	}

	return parsed, nil
}

// ParsedRetryConfig is the parsed version of RetryConfig
type ParsedRetryConfig struct {
	Attempts     int
	Backoff      BackoffStrategy
	InitialDelay time.Duration
	MaxDelay     time.Duration
	RetryOn      []int
	RetryOnError bool
}

// CalculateDelay calculates the delay for a given attempt number
func (p *ParsedRetryConfig) CalculateDelay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}

	var delay time.Duration

	switch p.Backoff {
	case BackoffConstant:
		delay = p.InitialDelay

	case BackoffLinear:
		delay = p.InitialDelay * time.Duration(attempt)

	case BackoffExponential:
		// 2^(attempt-1) * initialDelay
		multiplier := math.Pow(2, float64(attempt-1))
		delay = time.Duration(float64(p.InitialDelay) * multiplier)
	}

	// Cap at max delay
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay
}

// ShouldRetry determines if a retry should be attempted
func (p *ParsedRetryConfig) ShouldRetry(attempt int, statusCode int, err error) bool {
	// Check if we've exhausted attempts
	if attempt >= p.Attempts {
		return false
	}

	// If there's an error and RetryOnError is true, retry
	if err != nil && p.RetryOnError {
		return true
	}

	// If no specific status codes to retry on, don't retry on status alone
	if len(p.RetryOn) == 0 {
		return false
	}

	// Check if status code is in the retry list
	for _, code := range p.RetryOn {
		if statusCode == code {
			return true
		}
	}

	return false
}

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Attempts      int
	Success       bool
	LastError     error
	LastStatus    int
	TotalDuration time.Duration
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() (statusCode int, err error)

// ExecuteWithRetry executes a function with retry logic
func ExecuteWithRetry(config *ParsedRetryConfig, fn RetryableFunc) RetryResult {
	result := RetryResult{
		Attempts: 0,
		Success:  false,
	}

	startTime := time.Now()

	for attempt := 1; attempt <= config.Attempts; attempt++ {
		result.Attempts = attempt

		// Execute the function
		statusCode, err := fn()
		result.LastStatus = statusCode
		result.LastError = err

		// Check if successful (no error and not a retry status)
		if err == nil && statusCode < 400 && !config.shouldRetryStatus(statusCode) {
			result.Success = true
			result.TotalDuration = time.Since(startTime)
			return result
		}

		// Check if we should retry
		if !config.ShouldRetry(attempt, statusCode, err) {
			result.TotalDuration = time.Since(startTime)
			return result
		}

		// If not the last attempt, wait before retrying
		if attempt < config.Attempts {
			delay := config.CalculateDelay(attempt)
			time.Sleep(delay)
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result
}

// shouldRetryStatus checks if a status code should trigger a retry
func (p *ParsedRetryConfig) shouldRetryStatus(statusCode int) bool {
	if len(p.RetryOn) == 0 {
		return false
	}

	for _, code := range p.RetryOn {
		if statusCode == code {
			return true
		}
	}

	return false
}

// FormatRetryResult formats a retry result for logging
func FormatRetryResult(result RetryResult) string {
	if result.Success {
		if result.Attempts == 1 {
			return "success on first attempt"
		}
		return fmt.Sprintf("success after %d attempts (duration: %s)", result.Attempts, result.TotalDuration)
	}

	if result.LastError != nil {
		return fmt.Sprintf("failed after %d attempts: %v (duration: %s)", result.Attempts, result.LastError, result.TotalDuration)
	}

	return fmt.Sprintf("failed after %d attempts with status %d (duration: %s)", result.Attempts, result.LastStatus, result.TotalDuration)
}
