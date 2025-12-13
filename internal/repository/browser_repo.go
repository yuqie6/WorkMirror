package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yuqie6/mirror/internal/model"
	"gorm.io/gorm"
)

// BrowserEventRepository 浏览器事件仓储
type BrowserEventRepository struct {
	db *gorm.DB
}

// NewBrowserEventRepository 创建仓储
func NewBrowserEventRepository(db *gorm.DB) *BrowserEventRepository {
	return &BrowserEventRepository{db: db}
}

// Create 创建记录
func (r *BrowserEventRepository) Create(ctx context.Context, event *model.BrowserEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("创建浏览器事件失败: %w", err)
	}
	return nil
}

// BatchInsert 批量插入
func (r *BrowserEventRepository) BatchInsert(ctx context.Context, events []*model.BrowserEvent) error {
	if len(events) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(events, 100).Error; err != nil {
		return fmt.Errorf("批量插入浏览器事件失败: %w", err)
	}

	slog.Debug("批量插入浏览器事件", "count", len(events))
	return nil
}

// GetByDate 按日期查询
func (r *BrowserEventRepository) GetByDate(ctx context.Context, date string) ([]model.BrowserEvent, error) {
	startTime, endTime, err := DayRange(date)
	if err != nil {
		return nil, err
	}

	var events []model.BrowserEvent
	if err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp DESC").
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("查询浏览器事件失败: %w", err)
	}

	return events, nil
}

// GetByTimeRange 按时间范围查询浏览器事件
func (r *BrowserEventRepository) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]model.BrowserEvent, error) {
	var events []model.BrowserEvent
	if err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC").
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("查询浏览器事件失败: %w", err)
	}
	return events, nil
}

// GetByIDs 按 ID 列表批量查询浏览器事件（保持输入顺序）
func (r *BrowserEventRepository) GetByIDs(ctx context.Context, ids []int64) ([]model.BrowserEvent, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var events []model.BrowserEvent
	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("查询浏览器事件失败: %w", err)
	}

	byID := make(map[int64]model.BrowserEvent, len(events))
	for _, e := range events {
		byID[e.ID] = e
	}

	ordered := make([]model.BrowserEvent, 0, len(events))
	for _, id := range ids {
		if e, ok := byID[id]; ok {
			ordered = append(ordered, e)
		}
	}
	return ordered, nil
}

// GetDomainStats 获取域名统计
func (r *BrowserEventRepository) GetDomainStats(ctx context.Context, startTime, endTime int64, limit int) ([]DomainStat, error) {
	var stats []DomainStat
	if err := r.db.WithContext(ctx).
		Model(&model.BrowserEvent{}).
		Select("domain, COUNT(*) as visit_count, SUM(duration) as total_duration").
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Group("domain").
		Order("visit_count DESC").
		Limit(limit).
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("查询域名统计失败: %w", err)
	}

	return stats, nil
}

// DomainStat 域名统计
type DomainStat struct {
	Domain        string `json:"domain"`
	VisitCount    int64  `json:"visit_count"`
	TotalDuration int    `json:"total_duration"`
}

// CountByDateRange 统计日期范围内的事件数量
func (r *BrowserEventRepository) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.BrowserEvent{}).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计浏览器事件失败: %w", err)
	}
	return count, nil
}
