package service

import (
	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
)

const DefaultTopAppsLimit = 8

// SecondsToMinutesFloor 将秒数转换为分钟数（向下取整）
func SecondsToMinutesFloor(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return seconds / 60
}

// TopAppStats 返回 TopN 应用统计
func TopAppStats(stats []repository.AppStat, limit int) []repository.AppStat {
	if limit <= 0 || limit >= len(stats) {
		return stats
	}
	return stats[:limit]
}

// WindowEventInfosFromAppStats 将应用统计转换为 AI 请求用的窗口事件信息
func WindowEventInfosFromAppStats(stats []repository.AppStat, limit int) []ai.WindowEventInfo {
	picked := TopAppStats(stats, limit)
	out := make([]ai.WindowEventInfo, 0, len(picked))
	for _, stat := range picked {
		out = append(out, ai.WindowEventInfo{
			AppName:  stat.AppName,
			Duration: SecondsToMinutesFloor(stat.TotalDuration),
		})
	}
	return out
}

// SumCodingMinutesFromAppStats 统计代码编辑器的总使用时长（分钟）
func SumCodingMinutesFromAppStats(stats []repository.AppStat) int64 {
	var total int64
	for _, stat := range stats {
		if IsCodeEditor(stat.AppName) {
			total += int64(SecondsToMinutesFloor(stat.TotalDuration))
		}
	}
	return total
}
