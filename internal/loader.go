package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadTestCases(dir string) ([]TestCase, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}

	var all []TestCase
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var tcs []TestCase
		if err := yaml.Unmarshal(data, &tcs); err != nil {
			return nil, err
		}

		all = append(all, tcs...)
	}

	return all, nil
}
