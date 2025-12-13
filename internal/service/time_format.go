package service

import "time"

func FormatTimeRangeMs(startMs, endMs int64) string {
	if startMs <= 0 || endMs <= 0 || endMs <= startMs {
		return ""
	}
	start := time.UnixMilli(startMs).Format("15:04")
	end := time.UnixMilli(endMs).Format("15:04")
	return start + "-" + end
}
