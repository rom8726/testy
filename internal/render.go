package internal

import (
	"fmt"
	"regexp"
	"strconv"
)

var placeholderRe = regexp.MustCompile(`\{\{([a-zA-Z0-9\.\[\]_-]+)\}\}`)

func RenderTemplate(input string, ctx map[string]any) string {
	return placeholderRe.ReplaceAllStringFunc(input, func(match string) string {
		key := placeholderRe.FindStringSubmatch(match)[1]
		if val, ok := ctx[key]; ok {
			return fmt.Sprintf("%+v", val)
		}

		return match
	})
}

// tryConvertToNumber attempts to convert a string to a number if possible
func tryConvertToNumber(s string) any {
	// Try integer first
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Try boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return s
}

func RenderAny(v any, ctx map[string]any) any {
	switch val := v.(type) {
	case string:
		rendered := RenderTemplate(val, ctx)
		// If the original string was a placeholder and it got replaced,
		// try to convert the result to a number if it looks like one
		if placeholderRe.MatchString(val) && rendered != val {
			// The placeholder was replaced, try to convert to number
			return tryConvertToNumber(rendered)
		}
		return rendered
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = RenderAny(v2, ctx)
		}

		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = RenderAny(v2, ctx)
		}

		return out
	default:
		return val
	}
}

func renderRequest(req RequestSpec, ctx map[string]any) RequestSpec {
	req.Path = RenderTemplate(req.Path, ctx)
	req.Body = RenderAny(req.Body, ctx)

	for k, v := range req.Headers {
		req.Headers[k] = RenderTemplate(v, ctx)
	}

	return req
}

func renderDBCheck(check DBCheck, ctx map[string]any) DBCheck {
	check.Query = RenderTemplate(check.Query, ctx)
	check.Result = RenderAny(check.Result, ctx)

	return check
}
