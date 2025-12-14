package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/yuqie6/WorkMirror/internal/schema"
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
func (r *DiffRepository) Create(ctx context.Context, diff *schema.Diff) error {
	if err := r.db.WithContext(ctx).Create(diff).Error; err != nil {
		return fmt.Errorf("创建 Diff 记录失败: %w", err)
	}
	slog.Debug("Diff 记录已保存", "file", diff.FileName, "language", diff.Language)
	return nil
}

// GetByDate 按日期查询 Diff
func (r *DiffRepository) GetByDate(ctx context.Context, date string) ([]schema.Diff, error) {
	startTime, endTime, err := DayRange(date)
	if err != nil {
		return nil, err
	}

	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	return diffs, nil
}

// GetByTimeRange 按时间范围查询 Diff
func (r *DiffRepository) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Diff, error) {
	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Order("timestamp ASC").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}
	return diffs, nil
}

// GetByIDs 按 ID 列表批量查询 Diff（保持输入顺序）
func (r *DiffRepository) GetByIDs(ctx context.Context, ids []int64) ([]schema.Diff, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	byID := make(map[int64]schema.Diff, len(diffs))
	for _, d := range diffs {
		byID[d.ID] = d
	}

	ordered := make([]schema.Diff, 0, len(diffs))
	for _, id := range ids {
		if d, ok := byID[id]; ok {
			ordered = append(ordered, d)
		}
	}
	return ordered, nil
}

// GetByFilePath 按文件路径查询
func (r *DiffRepository) GetByFilePath(ctx context.Context, filePath string, limit int) ([]schema.Diff, error) {
	var diffs []schema.Diff
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
func (r *DiffRepository) GetByLanguage(ctx context.Context, language string, startTime, endTime int64) ([]schema.Diff, error) {
	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("language = ? AND timestamp >= ? AND timestamp <= ?", language, startTime, endTime).
		Order("timestamp DESC").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}

	return diffs, nil
}

// GetPendingAIAnalysis 获取待 AI 分析的 Diff
func (r *DiffRepository) GetPendingAIAnalysis(ctx context.Context, limit int) ([]schema.Diff, error) {
	var diffs []schema.Diff
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
		"skills_detected": schema.JSONArray(skills),
	}

	if err := r.db.WithContext(ctx).Model(&schema.Diff{}).
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
		Model(&schema.Diff{}).
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
	Language     string
	DiffCount    int64
	LinesAdded   int64
	LinesDeleted int64
}

// CountByDateRange 统计日期范围内的 Diff 数量
func (r *DiffRepository) CountByDateRange(ctx context.Context, startTime, endTime int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&schema.Diff{}).
		Where("timestamp >= ? AND timestamp <= ?", startTime, endTime).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计 Diff 数量失败: %w", err)
	}
	return count, nil
}

// GetLatestTimestamp 获取最新 Diff 时间戳（毫秒，无记录返回 0）
func (r *DiffRepository) GetLatestTimestamp(ctx context.Context) (int64, error) {
	var ts int64
	if err := r.db.WithContext(ctx).Model(&schema.Diff{}).
		Select("COALESCE(MAX(timestamp), 0)").
		Scan(&ts).Error; err != nil {
		return 0, fmt.Errorf("查询最新 Diff 时间失败: %w", err)
	}
	return ts, nil
}

// GetAllAnalyzed 获取所有已分析的 Diff
func (r *DiffRepository) GetAllAnalyzed(ctx context.Context) ([]schema.Diff, error) {
	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("ai_insight != '' AND ai_insight IS NOT NULL").
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询已分析 Diff 失败: %w", err)
	}
	return diffs, nil
}

// GetRecentAnalyzed 获取最近已分析的 Diff（按时间倒序）
func (r *DiffRepository) GetRecentAnalyzed(ctx context.Context, limit int) ([]schema.Diff, error) {
	if limit <= 0 {
		limit = 200
	}
	var diffs []schema.Diff
	if err := r.db.WithContext(ctx).
		Where("ai_insight != '' AND ai_insight IS NOT NULL").
		Order("timestamp DESC").
		Limit(limit).
		Find(&diffs).Error; err != nil {
		return nil, fmt.Errorf("查询最近 Diff 失败: %w", err)
	}
	return diffs, nil
}

// GetByID 根据 ID 查询 Diff
func (r *DiffRepository) GetByID(ctx context.Context, id int64) (*schema.Diff, error) {
	var diff schema.Diff
	if err := r.db.WithContext(ctx).First(&diff, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询 Diff 失败: %w", err)
	}
	return &diff, nil
}
