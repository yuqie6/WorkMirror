package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/repository"
	"github.com/yuqie6/mirror/internal/schema"
)

// SkillService 技能服务
type SkillService struct {
	skillRepo    SkillRepository
	diffRepo     DiffRepository
	activityRepo SkillActivityRepository
	expPolicy    ExpPolicy
}

// NewSkillService 创建技能服务
func NewSkillService(skillRepo SkillRepository, diffRepo DiffRepository, activityRepo SkillActivityRepository, expPolicy ExpPolicy) *SkillService {
	if expPolicy == nil {
		expPolicy = DefaultExpPolicy{}
	}
	return &SkillService{
		skillRepo:    skillRepo,
		diffRepo:     diffRepo,
		activityRepo: activityRepo,
		expPolicy:    expPolicy,
	}
}

// GetAllSkills 获取所有技能
func (s *SkillService) GetAllSkills(ctx context.Context) ([]schema.SkillNode, error) {
	return s.skillRepo.GetAll(ctx)
}

// ApplyContributions 统一入口：根据贡献更新技能
func (s *SkillService) ApplyContributions(ctx context.Context, contributions []SkillContribution) error {
	if len(contributions) == 0 {
		return nil
	}

	contributions = s.filterNewContributions(ctx, contributions)
	if len(contributions) == 0 {
		return nil
	}

	// 构建已存在技能名->Key 映射（用于父技能查表）
	existingSkills, _ := s.skillRepo.GetAll(ctx)
	nameToKeyMap := make(map[string]string)
	for _, sk := range existingSkills {
		nameToKeyMap[strings.ToLower(sk.Name)] = sk.Key
	}

	// 以 skill_key 聚合 exp 与最新元信息
	type agg struct {
		contrib SkillContribution
		expSum  float64
	}
	aggMap := make(map[string]*agg)
	for _, c := range contributions {
		if c.SkillKey == "" {
			c.SkillKey = normalizeKey(c.SkillName)
		}
		if c.SkillKey == "" {
			continue
		}
		a, ok := aggMap[c.SkillKey]
		if !ok {
			aggMap[c.SkillKey] = &agg{contrib: c, expSum: c.Exp}
			continue
		}
		a.expSum += c.Exp
		// 选择时间最新的元信息用于覆盖
		if c.Timestamp >= a.contrib.Timestamp {
			a.contrib = c
		}
	}

	skillsToUpdate := make([]*schema.SkillNode, 0, len(aggMap))
	for key, a := range aggMap {
		c := a.contrib

		// 父技能优先使用已存在的 Key，避免 Key 漂移
		parentKey := ""
		parentName := strings.TrimSpace(c.ParentName)
		if parentName != "" {
			if existingKey, ok := nameToKeyMap[strings.ToLower(parentName)]; ok {
				parentKey = existingKey
			} else {
				normalized := normalizeKey(parentName)
				if normalized != "" {
					if _, ok := nameToKeyMap[strings.ToLower(normalized)]; ok {
						parentKey = normalized
					} else if existing, _ := s.skillRepo.GetByKey(ctx, normalized); existing != nil {
						parentKey = normalized
					}
				}
			}
		}

		skill, err := s.skillRepo.GetByKey(ctx, key)
		if err != nil {
			slog.Warn("获取技能失败", "skill", key, "error", err)
			continue
		}

		if skill == nil {
			skill = schema.NewSkillNode(key, c.SkillName, c.Category)
			skill.ParentKey = parentKey
		} else {
			if c.SkillName != "" {
				skill.Name = c.SkillName
			}
			if c.Category != "" && c.Category != "other" {
				skill.Category = c.Category
			}
			if parentKey != "" {
				skill.ParentKey = parentKey
			}
		}

		if a.expSum > 0 {
			skill.AddExp(a.expSum)
		}
		skillsToUpdate = append(skillsToUpdate, skill)
	}

	return s.skillRepo.UpsertBatch(ctx, skillsToUpdate)
}

func (s *SkillService) filterNewContributions(ctx context.Context, contributions []SkillContribution) []SkillContribution {
	if s.activityRepo == nil {
		return contributions
	}

	seen := make(map[repository.SkillActivityKey]struct{}, len(contributions))
	keys := make([]repository.SkillActivityKey, 0, len(contributions))
	unique := make([]SkillContribution, 0, len(contributions))
	passthrough := make([]SkillContribution, 0)

	for _, c := range contributions {
		source := strings.TrimSpace(c.Source)
		if source == "" {
			source = "other"
		}
		key := c.SkillKey
		if key == "" {
			key = normalizeKey(c.SkillName)
		}
		c.Source = source
		c.SkillKey = key

		if key == "" || c.EvidenceID <= 0 || c.Timestamp <= 0 {
			// 无法做幂等：不写 activity，但仍允许技能更新
			passthrough = append(passthrough, c)
			continue
		}

		k := repository.SkillActivityKey{Source: source, EvidenceID: c.EvidenceID, SkillKey: key}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
		unique = append(unique, c)
	}

	existing, err := s.activityRepo.ListExistingKeys(ctx, keys)
	if err != nil {
		// 查询失败时不做过滤，保证核心流程可用
		return contributions
	}

	out := make([]SkillContribution, 0, len(unique)+len(passthrough))
	activities := make([]schema.SkillActivity, 0, len(keys))

	for _, c := range unique {
		k := repository.SkillActivityKey{Source: c.Source, EvidenceID: c.EvidenceID, SkillKey: c.SkillKey}
		if _, ok := existing[k]; ok {
			continue
		}
		out = append(out, c)
		activities = append(activities, schema.SkillActivity{
			SkillKey:   c.SkillKey,
			Source:     c.Source,
			EvidenceID: c.EvidenceID,
			Exp:        c.Exp,
			Timestamp:  c.Timestamp,
		})
	}
	out = append(out, passthrough...)

	if len(activities) > 0 {
		if _, err := s.activityRepo.BatchInsert(ctx, activities); err != nil {
			slog.Warn("写入技能活动失败", "error", err)
		}
	}

	return out
}

// UpdateSkillsFromDiffsWithCategory 根据 AI 返回的技能信息更新技能（兼容旧调用，内部转贡献）
func (s *SkillService) UpdateSkillsFromDiffsWithCategory(ctx context.Context, diffs []schema.Diff, skills []ai.SkillWithCategory) error {
	if len(skills) == 0 {
		return nil
	}

	totalLines := 0
	var latestTs int64
	var latestInsight string
	var latestFile string
	var latestDiffID int64
	for _, diff := range diffs {
		totalLines += diff.LinesAdded + diff.LinesDeleted
		if diff.Timestamp > latestTs {
			latestTs = diff.Timestamp
			latestInsight = diff.AIInsight
			latestFile = diff.FileName
			latestDiffID = diff.ID
		}
	}
	baseExp := s.expPolicy.CalcDiffExp(diffs)
	perSkillExp := baseExp / float64(len(skills))

	contribs := make([]SkillContribution, 0, len(skills))
	for _, aiSkill := range skills {
		skillKey := normalizeKey(aiSkill.Name)
		parentName := strings.TrimSpace(aiSkill.Parent)
		ctxText := latestInsight
		if ctxText == "" {
			ctxText = latestFile
		}
		contribs = append(contribs, SkillContribution{
			Source:              "diff",
			SkillKey:            skillKey,
			SkillName:           aiSkill.Name,
			Category:            aiSkill.Category,
			ParentName:          parentName,
			Exp:                 perSkillExp,
			EvidenceID:          latestDiffID,
			ContributionContext: ctxText,
			Timestamp:           latestTs,
		})
	}

	return s.ApplyContributions(ctx, contribs)
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
		skill = schema.NewSkillNode(skillKey, schema.NormalizeSkillName(skillKey), "other")
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

// SkillEvidence 技能证据（Phase B drill-down 最小返回）
type SkillEvidence struct {
	Source              string `json:"source"`
	EvidenceID          int64  `json:"evidence_id"`
	Timestamp           int64  `json:"timestamp"`
	ContributionContext string `json:"contribution_context"`
	FileName            string `json:"file_name,omitempty"`
	Insight             string `json:"insight,omitempty"`
}

// GetSkillEvidence 获取某技能最近的贡献证据（当前仅 Diff）
func (s *SkillService) GetSkillEvidence(ctx context.Context, skillKey string, limit int) ([]SkillEvidence, error) {
	if limit <= 0 {
		limit = 3
	}
	skill, err := s.skillRepo.GetByKey(ctx, skillKey)
	if err != nil || skill == nil {
		return nil, err
	}

	// 取最近已分析 diffs 并在内存中过滤
	diffs, err := s.diffRepo.GetRecentAnalyzed(ctx, 200)
	if err != nil {
		return nil, err
	}

	result := make([]SkillEvidence, 0, limit)
	seen := make(map[int64]struct{})
	for _, d := range diffs {
		if len(result) >= limit {
			break
		}
		if _, ok := seen[d.ID]; ok {
			continue
		}
		match := false
		for _, name := range d.SkillsDetected {
			if normalizeKey(name) == skillKey || strings.EqualFold(name, skill.Name) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		ctxText := strings.TrimSpace(d.AIInsight)
		if ctxText == "" {
			ctxText = d.FileName
		}
		ctxText = truncateRunes(ctxText, 120)
		result = append(result, SkillEvidence{
			Source:              "diff",
			EvidenceID:          d.ID,
			Timestamp:           d.Timestamp,
			ContributionContext: ctxText,
			FileName:            d.FileName,
			Insight:             d.AIInsight,
		})
		seen[d.ID] = struct{}{}
	}

	return result, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
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
func (s *SkillService) calculateTrend(skill *schema.SkillNode) string {
	daysInactive := skill.DaysInactive()
	if daysInactive == 0 {
		return "up"
	} else if daysInactive > 7 {
		return "down"
	}
	return "stable"
}
