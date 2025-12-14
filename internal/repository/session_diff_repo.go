package repository

import (
	"context"
	"fmt"

	"github.com/yuqie6/WorkMirror/internal/schema"
	"gorm.io/gorm"
)

// SessionDiffRepository 会话-Diff 关联仓储
type SessionDiffRepository struct {
	db *gorm.DB
}

// NewSessionDiffRepository 创建会话-Diff 关联仓储
func NewSessionDiffRepository(db *gorm.DB) *SessionDiffRepository {
	return &SessionDiffRepository{db: db}
}

// BatchInsert 批量插入关联
func (r *SessionDiffRepository) BatchInsert(ctx context.Context, sessionID int64, diffIDs []int64) error {
	if sessionID == 0 || len(diffIDs) == 0 {
		return nil
	}
	records := make([]schema.SessionDiff, 0, len(diffIDs))
	for _, id := range diffIDs {
		if id == 0 {
			continue
		}
		records = append(records, schema.SessionDiff{SessionID: sessionID, DiffID: id})
	}
	if len(records) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(records, 200).Error; err != nil {
		return fmt.Errorf("批量创建会话关联失败: %w", err)
	}
	return nil
}

// GetSessionIDsByDiffID 查询某 Diff 关联的会话
func (r *SessionDiffRepository) GetSessionIDsByDiffID(ctx context.Context, diffID int64) ([]int64, error) {
	var ids []int64
	if err := r.db.WithContext(ctx).
		Model(&schema.SessionDiff{}).
		Where("diff_id = ?", diffID).
		Pluck("session_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("查询会话关联失败: %w", err)
	}
	return ids, nil
}
