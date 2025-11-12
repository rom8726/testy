package internal

import (
	"errors"
	"testing"
	"time"
)

func TestParseRetryConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  RetryConfig
		wantErr bool
	}{
		{
			name: "valid config - exponential",
			config: RetryConfig{
				Attempts:     3,
				Backoff:      "exponential",
				InitialDelay: "100ms",
				MaxDelay:     "5s",
			},
			wantErr: false,
		},
		{
			name: "valid config - linear",
			config: RetryConfig{
				Attempts:     5,
				Backoff:      "linear",
				InitialDelay: "200ms",
			},
			wantErr: false,
		},
		{
			name: "valid config - constant",
			config: RetryConfig{
				Attempts:     2,
				Backoff:      "constant",
				InitialDelay: "1s",
			},
			wantErr: false,
		},
		{
			name: "default backoff",
			config: RetryConfig{
				Attempts: 3,
			},
			wantErr: false,
		},
		{
			name: "invalid backoff strategy",
			config: RetryConfig{
				Attempts: 3,
				Backoff:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid initial delay",
			config: RetryConfig{
				Attempts:     3,
				InitialDelay: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid max delay",
			config: RetryConfig{
				Attempts: 3,
				MaxDelay: "invalid",
			},
			wantErr: true,
		},
		{
			name: "zero attempts",
			config: RetryConfig{
				Attempts: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRetryConfig(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRetryConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalculateDelay_Constant(t *testing.T) {
	config := &ParsedRetryConfig{
		Backoff:      BackoffConstant,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 100 * time.Millisecond},
		{3, 100 * time.Millisecond},
		{10, 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := config.CalculateDelay(tt.attempt)
			if got != tt.want {
				t.Errorf("CalculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestCalculateDelay_Linear(t *testing.T) {
	config := &ParsedRetryConfig{
		Backoff:      BackoffLinear,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 300 * time.Millisecond},
		{10, 1 * time.Second}, // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := config.CalculateDelay(tt.attempt)
			if got != tt.want {
				t.Errorf("CalculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestCalculateDelay_Exponential(t *testing.T) {
	config := &ParsedRetryConfig{
		Backoff:      BackoffExponential,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 100 * time.Millisecond},  // 2^0 * 100ms
		{2, 200 * time.Millisecond},  // 2^1 * 100ms
		{3, 400 * time.Millisecond},  // 2^2 * 100ms
		{4, 800 * time.Millisecond},  // 2^3 * 100ms
		{5, 1600 * time.Millisecond}, // 2^4 * 100ms
		{6, 3200 * time.Millisecond}, // 2^5 * 100ms
		{7, 5 * time.Second},         // 2^6 * 100ms = 6400ms, capped at 5s
		{10, 5 * time.Second},        // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := config.CalculateDelay(tt.attempt)
			if got != tt.want {
				t.Errorf("CalculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		config     *ParsedRetryConfig
		attempt    int
		statusCode int
		err        error
		want       bool
	}{
		{
			name: "retry on error",
			config: &ParsedRetryConfig{
				Attempts:     3,
				RetryOnError: true,
			},
			attempt:    1,
			statusCode: 200,
			err:        errors.New("network error"),
			want:       true,
		},
		{
			name: "no retry when attempts exhausted",
			config: &ParsedRetryConfig{
				Attempts:     3,
				RetryOnError: true,
			},
			attempt:    3,
			statusCode: 200,
			err:        errors.New("network error"),
			want:       false,
		},
		{
			name: "retry on specific status code",
			config: &ParsedRetryConfig{
				Attempts: 3,
				RetryOn:  []int{503, 429},
			},
			attempt:    1,
			statusCode: 503,
			err:        nil,
			want:       true,
		},
		{
			name: "no retry on different status code",
			config: &ParsedRetryConfig{
				Attempts: 3,
				RetryOn:  []int{503, 429},
			},
			attempt:    1,
			statusCode: 500,
			err:        nil,
			want:       false,
		},
		{
			name: "no retry when no conditions match",
			config: &ParsedRetryConfig{
				Attempts:     3,
				RetryOnError: false,
			},
			attempt:    1,
			statusCode: 200,
			err:        nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ShouldRetry(tt.attempt, tt.statusCode, tt.err)
			if got != tt.want {
				t.Errorf("ShouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffConstant,
		InitialDelay: 10 * time.Millisecond,
		RetryOnError: true,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 200, nil
	}

	result := ExecuteWithRetry(config, fn)

	if !result.Success {
		t.Error("Expected success")
	}

	if result.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", result.Attempts)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d", callCount)
	}
}

func TestExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffConstant,
		InitialDelay: 10 * time.Millisecond,
		RetryOnError: true,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 3 {
			return 0, errors.New("temporary error")
		}
		return 200, nil
	}

	result := ExecuteWithRetry(config, fn)

	if !result.Success {
		t.Error("Expected success after retries")
	}

	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempts)
	}

	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, got %d", callCount)
	}
}

func TestExecuteWithRetry_FailureAfterAllAttempts(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffConstant,
		InitialDelay: 10 * time.Millisecond,
		RetryOnError: true,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 0, errors.New("persistent error")
	}

	result := ExecuteWithRetry(config, fn)

	if result.Success {
		t.Error("Expected failure")
	}

	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempts)
	}

	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, got %d", callCount)
	}

	if result.LastError == nil {
		t.Error("Expected LastError to be set")
	}
}

func TestExecuteWithRetry_RetryOnStatusCode(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffConstant,
		InitialDelay: 10 * time.Millisecond,
		RetryOn:      []int{503},
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		if callCount < 3 {
			return 503, nil
		}
		return 200, nil
	}

	result := ExecuteWithRetry(config, fn)

	if !result.Success {
		t.Error("Expected success after retries")
	}

	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempts)
	}
}

func TestExecuteWithRetry_NoRetryOn200(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffConstant,
		InitialDelay: 10 * time.Millisecond,
		RetryOn:      []int{503},
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 200, nil
	}

	result := ExecuteWithRetry(config, fn)

	if !result.Success {
		t.Error("Expected success")
	}

	if result.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", result.Attempts)
	}

	if callCount != 1 {
		t.Errorf("Expected function to be called once, got %d", callCount)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", config.Attempts)
	}

	if config.Backoff != "exponential" {
		t.Errorf("Expected exponential backoff, got %s", config.Backoff)
	}

	if config.InitialDelay != "100ms" {
		t.Errorf("Expected 100ms initial delay, got %s", config.InitialDelay)
	}

	if !config.RetryOnError {
		t.Error("Expected RetryOnError to be true")
	}
}

func TestFormatRetryResult(t *testing.T) {
	tests := []struct {
		name   string
		result RetryResult
		want   string
	}{
		{
			name: "success on first attempt",
			result: RetryResult{
				Attempts: 1,
				Success:  true,
			},
			want: "success on first attempt",
		},
		{
			name: "success after retries",
			result: RetryResult{
				Attempts:      3,
				Success:       true,
				TotalDuration: 500 * time.Millisecond,
			},
			want: "success after 3 attempts (duration: 500ms)",
		},
		{
			name: "failure with error",
			result: RetryResult{
				Attempts:      3,
				Success:       false,
				LastError:     errors.New("network timeout"),
				TotalDuration: 1 * time.Second,
			},
			want: "failed after 3 attempts: network timeout (duration: 1s)",
		},
		{
			name: "failure with status code",
			result: RetryResult{
				Attempts:      3,
				Success:       false,
				LastStatus:    503,
				TotalDuration: 1 * time.Second,
			},
			want: "failed after 3 attempts with status 503 (duration: 1s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRetryResult(tt.result)
			if got != tt.want {
				t.Errorf("FormatRetryResult() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecuteWithRetry_TimingExponential(t *testing.T) {
	config := &ParsedRetryConfig{
		Attempts:     3,
		Backoff:      BackoffExponential,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		RetryOnError: true,
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 0, errors.New("error")
	}

	start := time.Now()
	result := ExecuteWithRetry(config, fn)
	elapsed := time.Since(start)

	// Expected delays: 50ms + 100ms = 150ms minimum
	// With some tolerance for execution time
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected at least 100ms, got %v", elapsed)
	}

	if !result.Success {
		// This is expected
	}

	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempts)
	}
}
