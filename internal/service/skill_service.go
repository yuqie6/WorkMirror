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

// GetAllSkills 获取所有技能
func (s *SkillService) GetAllSkills(ctx context.Context) ([]model.SkillNode, error) {
	return s.skillRepo.GetAll(ctx)
}

// UpdateSkillsFromDiffsWithCategory 根据 AI 返回的技能信息更新技能（完全 AI 驱动）
func (s *SkillService) UpdateSkillsFromDiffsWithCategory(ctx context.Context, diffs []model.Diff, skills []ai.SkillWithCategory) error {
	if len(skills) == 0 {
		return nil
	}

	// 计算每个 diff 的基础经验
	totalLines := 0
	for _, diff := range diffs {
		totalLines += diff.LinesAdded + diff.LinesDeleted
	}
	baseExp := 1.0 + float64(totalLines)/10.0

	// 收集需要更新的技能
	skillsToUpdate := make([]*model.SkillNode, 0, len(skills))

	// 构建已存在技能名->Key 映射（用于父技能查表）
	existingSkills, _ := s.skillRepo.GetAll(ctx)
	nameToKeyMap := make(map[string]string)
	for _, sk := range existingSkills {
		nameToKeyMap[strings.ToLower(sk.Name)] = sk.Key
	}

	for _, aiSkill := range skills {
		// 统一 Key 格式
		skillKey := normalizeKey(aiSkill.Name)

		// 父技能优先使用已存在的 Key，避免 Key 漂移
		parentKey := ""
		if aiSkill.Parent != "" {
			if existingKey, ok := nameToKeyMap[strings.ToLower(aiSkill.Parent)]; ok {
				parentKey = existingKey
			} else {
				parentKey = normalizeKey(aiSkill.Parent)
			}
		}

		// 获取或创建技能
		skill, err := s.skillRepo.GetByKey(ctx, skillKey)
		if err != nil {
			slog.Warn("获取技能失败", "skill", skillKey, "error", err)
			continue
		}

		if skill == nil {
			// 创建新技能（使用 AI 决定的分类和父技能）
			skill = model.NewSkillNode(skillKey, aiSkill.Name, aiSkill.Category)
			skill.ParentKey = parentKey
		} else {
			// 更新分类和父技能（AI 优先）
			if aiSkill.Category != "" && aiSkill.Category != "other" {
				skill.Category = aiSkill.Category
			}
			if parentKey != "" {
				skill.ParentKey = parentKey
			}
		}

		// 添加经验
		skill.AddExp(baseExp / float64(len(skills))) // 均分经验

		skillsToUpdate = append(skillsToUpdate, skill)
	}

	// 批量更新
	if err := s.skillRepo.UpsertBatch(ctx, skillsToUpdate); err != nil {
		return err
	}

	return nil
}

// normalizeKey 统一 Key 格式（稳定 slug 策略）
func normalizeKey(name string) string {
	if name == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(name))

	// 常见特殊符号处理（在空格转换之前，按固定顺序替换，避免 map 遍历导致不稳定）
	orderedReplacements := []struct {
		old string
		new string
	}{
		// 更具体的后缀优先
		{"react.js", "reactjs"},
		{"vue.js", "vuejs"},
		{"next.js", "nextjs"},
		{"node.js", "nodejs"},
		// 语言/平台别名
		{"c++", "cpp"},
		{"c#", "csharp"},
		{".net", "dotnet"},
		// 通用后缀最后处理，避免误伤上面的替换结果
		{".js", "-js"},
		{".ts", "-ts"},
	}
	for _, rep := range orderedReplacements {
		key = strings.ReplaceAll(key, rep.old, rep.new)
	}

	// 空格转连字符
	key = strings.ReplaceAll(key, " ", "-")

	// 移除其他特殊字符（保留字母、数字、连字符）
	var result strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// addSkillExp 给技能添加经验（用于衰减恢复等场景）
func (s *SkillService) addSkillExp(ctx context.Context, skillKey string, exp float64) error {
	skill, err := s.skillRepo.GetByKey(ctx, skillKey)
	if err != nil {
		return err
	}

	if skill == nil {
		// 创建新技能（使用 key 作为名称，分类为 other）
		skill = model.NewSkillNode(skillKey, model.NormalizeSkillName(skillKey), "other")
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
			ParentKey:  skill.ParentKey,
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
	ParentKey  string    `json:"parent_key"` // 父技能 Key
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
