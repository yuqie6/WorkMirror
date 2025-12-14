package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yuqie6/mirror/internal/schema"
	"gorm.io/gorm"
)

// EventRepository 事件仓储
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository 创建事件仓储
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create 创建单个事件
func (r *EventRepository) Create(ctx context.Context, event *schema.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// BatchInsert 批量插入事件（事务包裹）
func (r *EventRepository) BatchInsert(ctx context.Context, events []schema.Event) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(events, 100).Error
	})

	if err != nil {
		slog.Error("批量插入事件失败", "count", len(events), "error", err)
		return fmt.Errorf("批量插入事件失败: %w", err)
	}

	slog.Debug("批量插入事件成功", "count", len(events), "duration", time.Since(start))
	return nil
}

// GetByTimeRange 按时间范围查询事件
func (r *EventRepository) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Event, error) {
	var events []schema.Event
	err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("查询事件失败: %w", err)
	}

	return events, nil
}

// GetByDate 按日期查询事件
func (r *EventRepository) GetByDate(ctx context.Context, date string) ([]schema.Event, error) {
	startTime, endTime, err := DayRange(date)
	if err != nil {
		return nil, err
	}
	return r.GetByTimeRange(ctx, startTime, endTime)
}

// GetByAppName 按应用名查询事件
func (r *EventRepository) GetByAppName(ctx context.Context, appName string, limit int) ([]schema.Event, error) {
	var events []schema.Event
	query := r.db.WithContext(ctx).Where("app_name = ?", appName).Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&events).Error; err != nil {
		return nil, fmt.Errorf("查询事件失败: %w", err)
	}

	return events, nil
}

// Count 统计事件总数
func (r *EventRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&schema.Event{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计事件失败: %w", err)
	}
	return count, nil
}

// CountByTimeRange 统计时间范围内的事件数量
func (r *EventRepository) CountByTimeRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&schema.Event{}).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计事件失败: %w", err)
	}
	return count, nil
}

// GetLatestTimestamp 获取最新事件时间戳（毫秒，无记录返回 0）
func (r *EventRepository) GetLatestTimestamp(ctx context.Context) (int64, error) {
	var ts int64
	if err := r.db.WithContext(ctx).Model(&schema.Event{}).
		Select("COALESCE(MAX(timestamp), 0)").
		Scan(&ts).Error; err != nil {
		return 0, fmt.Errorf("查询最新事件时间失败: %w", err)
	}
	return ts, nil
}

// GetAppStats 获取应用使用统计
func (r *EventRepository) GetAppStats(ctx context.Context, startTime, endTime int64) ([]AppStat, error) {
	var stats []AppStat
	err := r.db.WithContext(ctx).
		Model(&schema.Event{}).
		Select("app_name, SUM(duration) as total_duration, COUNT(*) as event_count").
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Group("app_name").
		Order("total_duration DESC").
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("查询应用统计失败: %w", err)
	}

	return stats, nil
}

// AppStat 应用使用统计
type AppStat struct {
	AppName       string
	TotalDuration int   // 总时长（秒）
	EventCount    int64 // 事件数
}

// DeleteOldEvents 删除旧事件（保留最近 N 天）
func (r *EventRepository) DeleteOldEvents(ctx context.Context, retainDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retainDays).UnixMilli()

	result := r.db.WithContext(ctx).
		Where("timestamp < ?", cutoffTime).
		Delete(&schema.Event{})

	if result.Error != nil {
		return 0, fmt.Errorf("删除旧事件失败: %w", result.Error)
	}

	slog.Info("清理旧事件", "deleted", result.RowsAffected, "retain_days", retainDays)
	return result.RowsAffected, nil
}
