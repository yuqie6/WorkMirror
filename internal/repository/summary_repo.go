package repository

import (
	"context"
	"fmt"

	"github.com/yuqie6/WorkMirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SummaryRepository 每日总结仓储
type SummaryRepository struct {
	db *gorm.DB
}

// NewSummaryRepository 创建仓储
func NewSummaryRepository(db *gorm.DB) *SummaryRepository {
	return &SummaryRepository{db: db}
}

// Upsert 插入或更新
func (r *SummaryRepository) Upsert(ctx context.Context, summary *schema.DailySummary) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}},
		UpdateAll: true,
	}).Create(summary).Error
}

// GetByDate 按日期获取
func (r *SummaryRepository) GetByDate(ctx context.Context, date string) (*schema.DailySummary, error) {
	var summary schema.DailySummary
	err := r.db.WithContext(ctx).Where("date = ?", date).First(&summary).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询总结失败: %w", err)
	}
	return &summary, nil
}

// GetRecent 获取最近的总结
func (r *SummaryRepository) GetRecent(ctx context.Context, days int) ([]schema.DailySummary, error) {
	var summaries []schema.DailySummary
	err := r.db.WithContext(ctx).
		Order("date DESC").
		Limit(days).
		Find(&summaries).Error
	if err != nil {
		return nil, fmt.Errorf("查询总结失败: %w", err)
	}
	return summaries, nil
}

// GetByDateRange 获取日期范围内的总结
func (r *SummaryRepository) GetByDateRange(ctx context.Context, startDate, endDate string) ([]schema.DailySummary, error) {
	var summaries []schema.DailySummary
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date DESC").
		Find(&summaries).Error
	if err != nil {
		return nil, fmt.Errorf("查询日期范围总结失败: %w", err)
	}
	return summaries, nil
}

// SummaryPreview 日报预览（用于历史列表）
type SummaryPreview struct {
	Date    string
	Preview string // 摘要前40字
}

// ListSummaryPreviews 获取已生成日报的预览列表
func (r *SummaryRepository) ListSummaryPreviews(ctx context.Context, limit int) ([]SummaryPreview, error) {
	var results []SummaryPreview
	err := r.db.WithContext(ctx).
		Model(&schema.DailySummary{}).
		Select("date, SUBSTR(summary, 1, 40) as preview").
		Order("date DESC").
		Limit(limit).
		Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询日报预览失败: %w", err)
	}
	return results, nil
}
