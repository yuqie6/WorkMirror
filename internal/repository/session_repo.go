package repository

import (
	"context"
	"fmt"

	"github.com/yuqie6/mirror/internal/schema"
	"gorm.io/gorm"
)

// SessionRepository 会话仓储
type SessionRepository struct {
	db *gorm.DB
}

const latestSessionVersionPerDateSQL = "session_version = (SELECT MAX(session_version) FROM sessions s2 WHERE s2.date = sessions.date)"

// NewSessionRepository 创建会话仓储
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 创建会话（尽量保持幂等：已存在则复用 ID）
func (r *SessionRepository) Create(ctx context.Context, session *schema.Session) (bool, error) {
	if session == nil {
		return false, fmt.Errorf("session is nil")
	}
	if session.StartTime <= 0 || session.EndTime <= 0 || session.EndTime <= session.StartTime {
		return false, fmt.Errorf("invalid session time range")
	}

	// 幂等保护：同一切分版本下，start/end 相同视为同一会话
	var existing schema.Session
	err := r.db.WithContext(ctx).
		Where("start_time = ? AND end_time = ? AND session_version = ?", session.StartTime, session.EndTime, session.SessionVersion).
		First(&existing).Error
	if err == nil {
		session.ID = existing.ID
		return false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("查询会话失败: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return false, fmt.Errorf("创建会话失败: %w", err)
	}
	return true, nil
}

// UpdateSemantic 更新会话语义字段（允许部分更新）
func (r *SessionRepository) UpdateSemantic(ctx context.Context, id int64, update schema.SessionSemanticUpdate) error {
	updates := map[string]interface{}{}
	if update.TimeRange != "" {
		updates["time_range"] = update.TimeRange
	}
	if update.Category != "" {
		updates["category"] = update.Category
	}
	if update.Summary != "" {
		updates["summary"] = update.Summary
	}
	if len(update.SkillsInvolved) > 0 {
		updates["skills_involved"] = schema.JSONArray(update.SkillsInvolved)
	}
	if update.Metadata != nil {
		updates["metadata"] = update.Metadata
	}
	if len(updates) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&schema.Session{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新会话语义失败: %w", err)
	}
	return nil
}

// GetByDate 按日期查询会话
func (r *SessionRepository) GetByDate(ctx context.Context, date string) ([]schema.Session, error) {
	startTime, endTime, err := DayRange(date)
	if err != nil {
		return nil, err
	}
	return r.GetByTimeRange(ctx, startTime, endTime)
}

// GetByTimeRange 按时间范围查询会话
func (r *SessionRepository) GetByTimeRange(ctx context.Context, startTime, endTime int64) ([]schema.Session, error) {
	var sessions []schema.Session
	if err := r.db.WithContext(ctx).
		Where("start_time >= ? AND start_time <= ?", startTime, endTime).
		Where(latestSessionVersionPerDateSQL).
		Order("start_time ASC").
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}
	return sessions, nil
}

// GetMaxSessionVersionByDate 获取某日期的最大切分版本号（无记录返回 0）
func (r *SessionRepository) GetMaxSessionVersionByDate(ctx context.Context, date string) (int, error) {
	var max int
	if err := r.db.WithContext(ctx).
		Model(&schema.Session{}).
		Select("COALESCE(MAX(session_version), 0)").
		Where("date = ?", date).
		Scan(&max).Error; err != nil {
		return 0, fmt.Errorf("查询会话版本失败: %w", err)
	}
	return max, nil
}

// GetByID 按 ID 查询会话
func (r *SessionRepository) GetByID(ctx context.Context, id int64) (*schema.Session, error) {
	var session schema.Session
	if err := r.db.WithContext(ctx).First(&session, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}
	return &session, nil
}

// GetLastSession 获取最近一次会话（按 end_time）
func (r *SessionRepository) GetLastSession(ctx context.Context) (*schema.Session, error) {
	var session schema.Session
	err := r.db.WithContext(ctx).
		Where(latestSessionVersionPerDateSQL).
		Order("end_time DESC").
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询最近会话失败: %w", err)
	}
	return &session, nil
}
