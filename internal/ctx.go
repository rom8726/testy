package internal

import (
	"fmt"
	"os"
	"strings"
)

func initCtxMap() map[string]any {
	envs := os.Environ()
	ctx := make(map[string]any, len(envs))

	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		ctx[parts[0]] = parts[1]
	}

	return ctx
}

func extractJSONFields(prefix string, data any, ctx map[string]any) {
	switch val := data.(type) {
	case map[string]any:
		for k, v := range val {
			key := fmt.Sprintf("%s.%s", prefix, k)
			extractJSONFields(key, v, ctx)
		}
	case []any:
		for i, v := range val {
			key := fmt.Sprintf("%s[%d]", prefix, i)
			extractJSONFields(key, v, ctx)
		}
	default:
		ctx[prefix] = val
	}
}
