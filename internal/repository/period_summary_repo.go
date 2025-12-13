package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/yuqie6/mirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PeriodSummaryRepository 阶段汇总仓储
type PeriodSummaryRepository struct {
	db *gorm.DB
}

// NewPeriodSummaryRepository 创建仓储
func NewPeriodSummaryRepository(db *gorm.DB) *PeriodSummaryRepository {
	return &PeriodSummaryRepository{db: db}
}

// Upsert 插入或更新
func (r *PeriodSummaryRepository) Upsert(ctx context.Context, summary *schema.PeriodSummary) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "start_date"}, {Name: "end_date"}},
		UpdateAll: true,
	}).Create(summary).Error
}

// GetByTypeAndRange 按类型和日期范围获取（带缓存时效检查）
func (r *PeriodSummaryRepository) GetByTypeAndRange(ctx context.Context, periodType, startDate, endDate string, maxAge time.Duration) (*schema.PeriodSummary, error) {
	var summary schema.PeriodSummary
	err := r.db.WithContext(ctx).
		Where("type = ? AND start_date = ? AND end_date = ?", periodType, startDate, endDate).
		First(&summary).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询阶段汇总失败: %w", err)
	}

	// 检查缓存是否过期
	if time.Since(summary.UpdatedAt) > maxAge {
		return nil, nil // 过期，返回 nil 触发重新生成
	}

	return &summary, nil
}

// ListByType 按类型获取历史汇总（按开始日期倒序）
func (r *PeriodSummaryRepository) ListByType(ctx context.Context, periodType string, limit int) ([]schema.PeriodSummary, error) {
	var summaries []schema.PeriodSummary
	err := r.db.WithContext(ctx).
		Where("type = ?", periodType).
		Order("start_date DESC").
		Limit(limit).
		Find(&summaries).Error
	if err != nil {
		return nil, fmt.Errorf("查询历史汇总失败: %w", err)
	}
	return summaries, nil
}
