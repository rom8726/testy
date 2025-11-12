package internal

import "fmt"

// RenderMode defines how to handle missing placeholders
type RenderMode int

const (
	// RenderModePermissive allows missing placeholders to remain unchanged
	RenderModePermissive RenderMode = iota
	// RenderModeStrict fails on missing placeholders
	RenderModeStrict
)

// RenderOptions configures placeholder rendering behavior
type RenderOptions struct {
	Mode         RenderMode
	DefaultValue string // used when Mode is Permissive and placeholder is missing
}

// DefaultRenderOptions returns the default rendering options
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		Mode:         RenderModePermissive,
		DefaultValue: "",
	}
}

// RenderTemplateWithOptions renders a template with the given options
func RenderTemplateWithOptions(input string, ctx map[string]any, opts RenderOptions) (string, error) {
	var lastErr error

	result := placeholderRe.ReplaceAllStringFunc(input, func(match string) string {
		key := placeholderRe.FindStringSubmatch(match)[1]
		if val, ok := ctx[key]; ok {
			return fmt.Sprintf("%+v", val)
		}

		// Handle missing placeholder based on mode
		if opts.Mode == RenderModeStrict {
			lastErr = fmt.Errorf("placeholder not found in context: %s", key)

			return match
		}

		// Permissive mode
		if opts.DefaultValue != "" {
			return opts.DefaultValue
		}

		return match
	})

	return result, lastErr
}

// RenderAnyWithOptions renders any value with the given options
func RenderAnyWithOptions(v any, ctx map[string]any, opts RenderOptions) (any, error) {
	switch val := v.(type) {
	case string:
		return RenderTemplateWithOptions(val, ctx, opts)
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			rendered, err := RenderAnyWithOptions(v2, ctx, opts)
			if err != nil {
				return nil, err
			}
			out[k] = rendered
		}
		return out, nil
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			rendered, err := RenderAnyWithOptions(v2, ctx, opts)
			if err != nil {
				return nil, err
			}
			out[i] = rendered
		}
		return out, nil
	default:
		return val, nil
	}
}
