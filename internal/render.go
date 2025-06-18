package internal

import (
	"fmt"
	"regexp"
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

func RenderAny(v any, ctx map[string]any) any {
	switch val := v.(type) {
	case string:
		return RenderTemplate(val, ctx)
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

	return check
}
