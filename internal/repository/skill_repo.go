package repository

import (
	"context"
	"fmt"

	"github.com/yuqie6/mirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SkillRepository 技能仓储
type SkillRepository struct {
	db *gorm.DB
}

// NewSkillRepository 创建仓储
func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

// GetByKey 根据 Key 获取技能
func (r *SkillRepository) GetByKey(ctx context.Context, key string) (*schema.SkillNode, error) {
	var skill schema.SkillNode
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&skill).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}
	return &skill, nil
}

// Upsert 插入或更新技能
func (r *SkillRepository) Upsert(ctx context.Context, skill *schema.SkillNode) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(skill).Error
}

// GetAll 获取所有技能
func (r *SkillRepository) GetAll(ctx context.Context) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).Order("level DESC, exp DESC").Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}
	return skills, nil
}

// GetByCategory 根据分类获取技能
func (r *SkillRepository) GetByCategory(ctx context.Context, category string) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("level DESC, exp DESC").
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}
	return skills, nil
}

// GetTopSkills 获取排名前 N 的技能
func (r *SkillRepository) GetTopSkills(ctx context.Context, limit int) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Order("level DESC, exp DESC").
		Limit(limit).
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}
	return skills, nil
}

// GetRecentlyActive 获取最近活跃的技能
func (r *SkillRepository) GetRecentlyActive(ctx context.Context, limit int) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Order("last_active DESC").
		Limit(limit).
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询技能失败: %w", err)
	}
	return skills, nil
}

// Count 统计技能数量
func (r *SkillRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&schema.SkillNode{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("统计技能失败: %w", err)
	}
	return count, nil
}

// Transaction 在事务中执行操作
func (r *SkillRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// UpsertBatch 批量插入或更新技能（在事务中）
func (r *SkillRepository) UpsertBatch(ctx context.Context, skills []*schema.SkillNode) error {
	if len(skills) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, skill := range skills {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				UpdateAll: true,
			}).Create(skill).Error; err != nil {
				return fmt.Errorf("批量更新技能失败: %w", err)
			}
		}
		return nil
	})
}

// GetByParent 获取某个技能的所有子技能
func (r *SkillRepository) GetByParent(ctx context.Context, parentKey string) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Where("parent_key = ?", parentKey).
		Order("level DESC, exp DESC").
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询子技能失败: %w", err)
	}
	return skills, nil
}

// GetTopLevel 获取所有顶级技能（没有父技能）
func (r *SkillRepository) GetTopLevel(ctx context.Context) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Where("parent_key = '' OR parent_key IS NULL").
		Order("category, level DESC, exp DESC").
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询顶级技能失败: %w", err)
	}
	return skills, nil
}

// GetActiveSkillsInPeriod 获取指定时间窗内活跃的技能（按最近活跃时间过滤）
func (r *SkillRepository) GetActiveSkillsInPeriod(ctx context.Context, startTime, endTime int64, limit int) ([]schema.SkillNode, error) {
	var skills []schema.SkillNode
	err := r.db.WithContext(ctx).
		Where("last_active >= ? AND last_active <= ?", startTime, endTime).
		Order("level DESC, exp DESC").
		Limit(limit).
		Find(&skills).Error
	if err != nil {
		return nil, fmt.Errorf("查询活跃技能失败: %w", err)
	}
	return skills, nil
}
