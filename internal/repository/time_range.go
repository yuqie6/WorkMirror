package repository

import (
	"fmt"
	"time"
)

// DayRange 将 YYYY-MM-DD 解析为本地日区间的毫秒时间戳 [start, end]（闭区间）。
func DayRange(date string) (startMs int64, endMs int64, err error) {
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return 0, 0, fmt.Errorf("解析日期失败: %w", err)
	}
	start := t.UnixMilli()
	end := t.Add(24*time.Hour).UnixMilli() - 1
	return start, end, nil
}
