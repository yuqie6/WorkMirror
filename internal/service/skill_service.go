package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/repository"
)

// SkillService 技能服务
type SkillService struct {
	skillRepo *repository.SkillRepository
	diffRepo  *repository.DiffRepository
}

// NewSkillService 创建技能服务
func NewSkillService(skillRepo *repository.SkillRepository, diffRepo *repository.DiffRepository) *SkillService {
	return &SkillService{
		skillRepo: skillRepo,
		diffRepo:  diffRepo,
	}
}

// UpdateSkillsFromDiffs 根据 Diff 更新技能
func (s *SkillService) UpdateSkillsFromDiffs(ctx context.Context, diffs []model.Diff) error {
	skillExp := make(map[string]float64) // skill key -> exp to add

	for _, diff := range diffs {
		// 根据语言添加基础经验
		langKey := s.getLanguageSkillKey(diff.Language)
		if langKey != "" {
			// 基础经验：每次修改 +1，每 10 行变更额外 +1
			baseExp := 1.0 + float64(diff.LinesAdded+diff.LinesDeleted)/10.0
			skillExp[langKey] += baseExp
		}

		// 根据 AI 检测的技能添加经验
		for _, skill := range diff.SkillsDetected {
			skillKey := s.normalizeSkillKey(skill)
			skillExp[skillKey] += 0.5 // 每个检测到的技能 +0.5
		}
	}

	// 更新技能
	for skillKey, exp := range skillExp {
		if err := s.addSkillExp(ctx, skillKey, exp); err != nil {
			slog.Warn("更新技能经验失败", "skill", skillKey, "error", err)
		}
	}

	return nil
}

// UpdateSkillsFromDiffsWithCategory 根据 Diff 和 AI 返回的技能分类更新技能(使用 AI 决定的分类)
func (s *SkillService) UpdateSkillsFromDiffsWithCategory(ctx context.Context, diffs []model.Diff, skills []ai.SkillWithCategory) error {
	skillExp := make(map[string]float64)     // skill key -> exp to add
	skillCategory := make(map[string]string) // skill key -> category

	for _, diff := range diffs {
		// 根据语言添加基础经验
		langKey := s.getLanguageSkillKey(diff.Language)
		if langKey != "" {
			baseExp := 1.0 + float64(diff.LinesAdded+diff.LinesDeleted)/10.0
			skillExp[langKey] += baseExp
			skillCategory[langKey] = "language"
		}
	}

	// 根据 AI 检测的技能添加经验（使用 AI 决定的分类）
	for _, skill := range skills {
		skillKey := s.normalizeSkillKey(skill.Name)
		skillExp[skillKey] += 0.5
		skillCategory[skillKey] = skill.Category
	}

	// 收集需要更新的技能
	skillsToUpdate := make([]*model.SkillNode, 0, len(skillExp))
	for skillKey, exp := range skillExp {
		category := skillCategory[skillKey]
		if category == "" {
			category = "other"
		}

		skill, err := s.skillRepo.GetByKey(ctx, skillKey)
		if err != nil {
			slog.Warn("获取技能失败", "skill", skillKey, "error", err)
			continue
		}

		if skill == nil {
			skill = model.NewSkillNode(skillKey, s.getSkillName(skillKey), category)
		} else if skill.Category != category && category != "" && category != "other" {
			skill.Category = category
		}

		skill.AddExp(exp)
		skillsToUpdate = append(skillsToUpdate, skill)
	}

	// 使用事务批量更新
	if err := s.skillRepo.UpsertBatch(ctx, skillsToUpdate); err != nil {
		return err
	}

	return nil
}

// addSkillExp 给技能添加经验
func (s *SkillService) addSkillExp(ctx context.Context, skillKey string, exp float64) error {
	skill, err := s.skillRepo.GetByKey(ctx, skillKey)
	if err != nil {
		return err
	}

	if skill == nil {
		// 创建新技能
		skill = model.NewSkillNode(skillKey, s.getSkillName(skillKey), s.getSkillCategory(skillKey))
	}

	// 添加经验
	skill.AddExp(exp)

	// 保存
	return s.skillRepo.Upsert(ctx, skill)
}

// ApplyDecayToAll 对所有技能应用衰减
func (s *SkillService) ApplyDecayToAll(ctx context.Context) error {
	skills, err := s.skillRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	decayed := 0
	for _, skill := range skills {
		if skill.DaysInactive() > 7 {
			oldExp := skill.Exp
			skill.ApplyDecay()
			if skill.Exp != oldExp {
				if err := s.skillRepo.Upsert(ctx, &skill); err != nil {
					slog.Warn("保存技能衰减失败", "skill", skill.Key, "error", err)
					continue
				}
				decayed++
			}
		}
	}

	if decayed > 0 {
		slog.Info("技能衰减已应用", "count", decayed)
	}

	return nil
}

// GetSkillTree 获取技能树
func (s *SkillService) GetSkillTree(ctx context.Context) (*SkillTree, error) {
	skills, err := s.skillRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	tree := &SkillTree{
		Categories: make(map[string][]SkillNodeView),
		UpdatedAt:  time.Now(),
	}

	for _, skill := range skills {
		view := SkillNodeView{
			Key:        skill.Key,
			Name:       skill.Name,
			Level:      skill.Level,
			Exp:        skill.Exp,
			ExpToNext:  skill.ExpToNext,
			Progress:   skill.Exp / skill.ExpToNext * 100,
			LastActive: time.UnixMilli(skill.LastActive),
			Trend:      s.calculateTrend(&skill),
		}

		category := skill.Category
		if category == "" {
			category = "other"
		}
		tree.Categories[category] = append(tree.Categories[category], view)
		tree.TotalSkills++
	}

	return tree, nil
}

// SkillTree 技能树视图
type SkillTree struct {
	Categories  map[string][]SkillNodeView `json:"categories"`
	TotalSkills int                        `json:"total_skills"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

// SkillNodeView 技能节点视图
type SkillNodeView struct {
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	Level      int       `json:"level"`
	Exp        float64   `json:"exp"`
	ExpToNext  float64   `json:"exp_to_next"`
	Progress   float64   `json:"progress"`
	LastActive time.Time `json:"last_active"`
	Trend      string    `json:"trend"` // up, down, stable
}

// calculateTrend 计算技能趋势
func (s *SkillService) calculateTrend(skill *model.SkillNode) string {
	daysInactive := skill.DaysInactive()
	if daysInactive == 0 {
		return "up"
	} else if daysInactive > 7 {
		return "down"
	}
	return "stable"
}

// getLanguageSkillKey 获取语言技能 Key
func (s *SkillService) getLanguageSkillKey(language string) string {
	if language == "" {
		return ""
	}
	// 直接使用语言名生成 key（AI 已经决定分类）
	return "lang." + strings.ToLower(language)
}

// normalizeSkillKey 标准化技能 Key
func (s *SkillService) normalizeSkillKey(skill string) string {
	// 简单处理：转小写，空格替换为点
	key := strings.ToLower(strings.TrimSpace(skill))
	key = strings.ReplaceAll(key, " ", ".")
	return "skill." + key
}

// getSkillName 获取技能显示名称
func (s *SkillService) getSkillName(skillKey string) string {
	// 从 key 提取名称并标准化
	parts := strings.Split(skillKey, ".")
	name := skillKey
	if len(parts) > 1 {
		name = parts[len(parts)-1]
	}
	return model.NormalizeSkillName(name)
}

// getSkillCategory 获取技能分类
func (s *SkillService) getSkillCategory(skillKey string) string {
	// 从 key 提取技能名并获取分类
	parts := strings.Split(skillKey, ".")
	name := skillKey
	if len(parts) > 1 {
		name = parts[len(parts)-1]
	}
	return string(model.GetSkillCategory(name))
}
