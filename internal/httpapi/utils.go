//go:build windows

package httpapi

import (
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/model"
)

func formatTimeRangeMs(startMs, endMs int64) string {
	if startMs <= 0 || endMs <= 0 || endMs <= startMs {
		return ""
	}
	start := time.UnixMilli(startMs).Format("15:04")
	end := time.UnixMilli(endMs).Format("15:04")
	return start + "-" + end
}

func toInt64Slice(meta model.JSONMap, key string) []int64 {
	if meta == nil {
		return nil
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []int64:
		return append([]int64(nil), v...)
	case []interface{}:
		out := make([]int64, 0, len(v))
		for _, it := range v {
			switch n := it.(type) {
			case int64:
				out = append(out, n)
			case int:
				out = append(out, int64(n))
			case float64:
				out = append(out, int64(n))
			case float32:
				out = append(out, int64(n))
			}
		}
		return out
	default:
		return nil
	}
}

func toStringSlice(meta model.JSONMap, key string) []string {
	if meta == nil {
		return nil
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		out := make([]string, 0, len(v))
		for _, s := range v {
			if strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func toMapSlice(meta model.JSONMap, key string) []map[string]any {
	if meta == nil {
		return nil
	}
	raw, ok := meta[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []map[string]any:
		return v
	case []interface{}:
		out := make([]map[string]any, 0, len(v))
		for _, it := range v {
			if m, ok := it.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}
