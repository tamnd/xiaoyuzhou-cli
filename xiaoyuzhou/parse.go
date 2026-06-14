package xiaoyuzhou

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

// getPageProps extracts pageProps from a __NEXT_DATA__ script tag.
func getPageProps(body []byte) (map[string]any, error) {
	m := nextDataRe.FindSubmatch(body)
	if len(m) < 2 {
		return nil, fmt.Errorf("__NEXT_DATA__ not found")
	}
	var root map[string]any
	if err := json.Unmarshal(m[1], &root); err != nil {
		return nil, fmt.Errorf("parse __NEXT_DATA__: %w", err)
	}
	props, _ := root["props"].(map[string]any)
	pageProps, _ := props["pageProps"].(map[string]any)
	if pageProps == nil {
		return nil, fmt.Errorf("pageProps not found")
	}
	return pageProps, nil
}

// strVal returns a nil-safe, trimmed string representation of v.
func strVal(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

// formatDuration converts a float64 second count to "m:ss". Returns "-" for
// zero or nil values.
func formatDuration(v any) string {
	var secs float64
	switch x := v.(type) {
	case float64:
		secs = x
	case int:
		secs = float64(x)
	case int64:
		secs = float64(x)
	}
	if secs <= 0 {
		return "-"
	}
	m := int(secs) / 60
	s := int(secs) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
