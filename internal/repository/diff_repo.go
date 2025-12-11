package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/danielsclee/mirror/internal/model"
	"gorm.io/gorm"
)

// DiffRepository Diff 仓储
type DiffRepository struct {
	db *gorm.DB
}

// NewDiffRepository 创建 Diff 仓储
func NewDiffRepository(db *gorm.DB) *DiffRepository {
	return &DiffRepository{db: db}
}

// Create 创建单个 Diff 记录
func (r *DiffRepository) Create(ctx context.Context, diff *model.Diff) error {
	if err := r.db.WithContext(ctx).Create(diff).Error; err != nil {
		return fmt.Errorf("创建 Diff 记录失败: %w", err)
	}
	slog.Debug("Diff 记录已保存", "file", diff.FileName, "language", diff.Language)
	return nil
}

// GetByDate 按日期查询 Diff
func (r *DiffRepository) GetByDate(ctx context.Context, date string) ([]model.Diff, error) {
	loc := time.Local
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("解析日期失败: %w", err)
	}

	startTime := t.UnixMilli()
	endTime := t.Add(24*time.Hour).UnixMilli() - 1

	var diffs []model.Diff
	if err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	return diffs, nil
}

// GetByFilePath 按文件路径查询
func (r *DiffRepository) GetByFilePath(ctx context.Context, filePath string, limit int) ([]model.Diff, error) {
	var diffs []model.Diff
	query := r.db.WithContext(ctx).Where("file_path = ?", filePath).Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	return diffs, nil
}

// GetByLanguage 按语言查询
func (r *DiffRepository) GetByLanguage(ctx context.Context, language string, startTime, endTime int64) ([]model.Diff, error) {
	var diffs []model.Diff
	if err := r.db.WithContext(ctx).
		Where("language = ? AND timestamp >= ? AND timestamp <= ?", language, startTime, endTime).
		Order("timestamp DESC").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	return diffs, nil
}

// GetPendingAIAnalysis 获取待 AI 分析的 Diff
func (r *DiffRepository) GetPendingAIAnalysis(ctx context.Context, limit int) ([]model.Diff, error) {
	var diffs []model.Diff
	if err := r.db.WithContext(ctx).
		Where("ai_insight = '' OR ai_insight IS NULL").
		Order("timestamp ASC").
		Limit(limit).
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询待分析 Diff 失败: %w", err)
	}

	return diffs, nil
}

// UpdateAIInsight 更新 AI 解读
func (r *DiffRepository) UpdateAIInsight(ctx context.Context, id int64, insight string, skills []string) error {
	updates := map[string]interface{}{
		"ai_insight":      insight,
		"skills_detected": model.JSONArray(skills),
	}

	if err := r.db.WithContext(ctx).Model(&model.Diff{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新 AI 解读失败: %w", err)
	}

	return nil
}

// GetLanguageStats 获取语言统计
func (r *DiffRepository) GetLanguageStats(ctx context.Context, startTime, endTime int64) ([]LanguageStat, error) {
	var stats []LanguageStat
	if err := r.db.WithContext(ctx).
		Model(&model.Diff{}).
		Select("language, COUNT(*) as diff_count, SUM(lines_added) as lines_added, SUM(lines_deleted) as lines_deleted").
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Group("language").
		Order("diff_count DESC").
		Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("查询语言统计失败: %w", err)
	}

	return stats, nil
}

// LanguageStat 语言统计
type LanguageStat struct {
	Language     string `json:"language"`
	DiffCount    int64  `json:"diff_count"`
	LinesAdded   int64  `json:"lines_added"`
	LinesDeleted int64  `json:"lines_deleted"`
}

// CountByDateRange 统计日期范围内的 Diff 数量
func (r *DiffRepository) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Diff{}).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计 Diff 数量失败: %w", err)
	}
	return count, nil
}

// GetAllAnalyzed 获取所有已分析的 Diff
func (r *DiffRepository) GetAllAnalyzed(ctx context.Context) ([]model.Diff, error) {
	var diffs []model.Diff
	if err := r.db.WithContext(ctx).
		Where("ai_insight != '' AND ai_insight IS NOT NULL").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询已分析 Diff 失败: %w", err)
	}
	return diffs, nil
}
