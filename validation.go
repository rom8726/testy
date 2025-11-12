package testy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func (c *Config) Validate() error {
	var validationErrors []ValidationError

	if c.Handler == nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Handler",
			Message: "http.Handler is required",
		})
	}

	if c.CasesDir == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "CasesDir",
			Message: "CasesDir is required",
		})
	} else {
		if info, err := os.Stat(c.CasesDir); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "CasesDir",
				Message: fmt.Sprintf("directory does not exist: %s", c.CasesDir),
			})
		} else if !info.IsDir() {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "CasesDir",
				Message: fmt.Sprintf("path is not a directory: %s", c.CasesDir),
			})
		}
	}

	if c.FixturesDir != "" {
		if info, err := os.Stat(c.FixturesDir); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "FixturesDir",
				Message: fmt.Sprintf("directory does not exist: %s", c.FixturesDir),
			})
		} else if !info.IsDir() {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "FixturesDir",
				Message: fmt.Sprintf("path is not a directory: %s", c.FixturesDir),
			})
		}
	}

	if c.FixturesDir != "" && c.ConnStr == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "ConnStr",
			Message: "database connection string is required when FixturesDir is provided",
		})
	}

	if c.ConnStr != "" && len(c.DBType) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "DBType",
			Message: "DBType must be specified when ConnStr is provided",
		})
	}

	if c.JUnitReport != "" {
		dir := filepath.Dir(c.JUnitReport)
		if dir != "" && dir != "." {
			if info, err := os.Stat(dir); err != nil {
				if err := os.MkdirAll(dir, 0755); err != nil {
					validationErrors = append(validationErrors, ValidationError{
						Field:   "JUnitReport",
						Message: fmt.Sprintf("cannot create report directory: %s", err),
					})
				}
			} else if !info.IsDir() {
				validationErrors = append(validationErrors, ValidationError{
					Field:   "JUnitReport",
					Message: fmt.Sprintf("parent path is not a directory: %s", dir),
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		var errMsg string
		for i, ve := range validationErrors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += ve.Error()
		}

		return errors.New(errMsg)
	}

	return nil
}
