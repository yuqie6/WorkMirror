package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/danielsclee/mirror/internal/model"
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
func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// BatchInsert 批量插入事件（事务包裹）
func (r *EventRepository) BatchInsert(ctx context.Context, events []model.Event) error {
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
func (r *EventRepository) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.Event, error) {
	var events []model.Event
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
func (r *EventRepository) GetByDate(ctx context.Context, date string) ([]model.Event, error) {
	// 解析日期为时间戳范围
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("解析日期失败: %w", err)
	}

	startTime := t.UnixMilli()
	endTime := t.Add(24*time.Hour).UnixMilli() - 1

	return r.GetByTimeRange(ctx, startTime, endTime)
}

// GetByAppName 按应用名查询事件
func (r *EventRepository) GetByAppName(ctx context.Context, appName string, limit int) ([]model.Event, error) {
	var events []model.Event
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
	if err := r.db.WithContext(ctx).Model(&model.Event{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计事件失败: %w", err)
	}
	return count, nil
}

// GetAppStats 获取应用使用统计
func (r *EventRepository) GetAppStats(ctx context.Context, startTime, endTime int64) ([]AppStat, error) {
	var stats []AppStat
	err := r.db.WithContext(ctx).
		Model(&model.Event{}).
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
	AppName       string `json:"app_name"`
	TotalDuration int    `json:"total_duration"` // 总时长（秒）
	EventCount    int64  `json:"event_count"`    // 事件数
}

// DeleteOldEvents 删除旧事件（保留最近 N 天）
func (r *EventRepository) DeleteOldEvents(ctx context.Context, retainDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retainDays).UnixMilli()

	result := r.db.WithContext(ctx).
		Where("timestamp < ?", cutoffTime).
		Delete(&model.Event{})

	if result.Error != nil {
		return 0, fmt.Errorf("删除旧事件失败: %w", result.Error)
	}

	slog.Info("清理旧事件", "deleted", result.RowsAffected, "retain_days", retainDays)
	return result.RowsAffected, nil
}
