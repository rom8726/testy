package loader

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/rom8726/testy/types"
)

func LoadTestCases(dir string) ([]types.TestCase, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}

	var all []types.TestCase
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var tcs []types.TestCase
		if err := yaml.Unmarshal(data, &tcs); err != nil {
			return nil, err
		}

		all = append(all, tcs...)
	}

	return all, nil
}
