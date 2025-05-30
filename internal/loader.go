package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadTestCases(dir string) ([]TestCase, error) {
	const op = "LoadTestCases"

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, NewError(ErrNotFound, op, "directory does not exist").
			WithContext("directory", dir)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, NewError(ErrInternal, op, "failed to find test case files").
			WithContext("directory", dir).
			WithContext("error", err.Error())
	}

	var all []TestCase
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, NewError(ErrNotFound, op, "failed to read test case file").
				WithContext("file", file).
				WithContext("error", err.Error())
		}

		var tcs []TestCase
		if err := yaml.Unmarshal(data, &tcs); err != nil {
			return nil, NewError(ErrInvalidInput, op, "failed to parse test case file").
				WithContext("file", file).
				WithContext("error", err.Error())
		}

		all = append(all, tcs...)
	}

	return all, nil
}
