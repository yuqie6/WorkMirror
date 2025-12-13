//go:build windows

package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/dto"
	"github.com/yuqie6/mirror/internal/schema"
	"github.com/yuqie6/mirror/internal/service"
)

func (a *API) HandleSkillTree(w http.ResponseWriter, r *http.Request) {
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		WriteError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	skillTree, err := a.rt.Core.Services.Skills.GetSkillTree(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []dto.SkillNodeDTO
	for category, skills := range skillTree.Categories {
		for _, skill := range skills {
			result = append(result, dto.SkillNodeDTO{
				Key:        skill.Key,
				Name:       skill.Name,
				Category:   category,
				ParentKey:  skill.ParentKey,
				Level:      skill.Level,
				Experience: int(skill.Exp),
				Progress:   int(skill.Progress),
				Status:     skill.Trend,
				LastActive: skill.LastActive.UnixMilli(),
			})
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleSkillEvidence(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		WriteError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.Skills == nil {
		WriteError(w, http.StatusBadRequest, "技能服务未初始化")
		return
	}
	evs, err := a.rt.Core.Services.Skills.GetSkillEvidence(r.Context(), skillKey, 3)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]dto.SkillEvidenceDTO, len(evs))
	for i, e := range evs {
		result[i] = dto.SkillEvidenceDTO{
			Source:              e.Source,
			EvidenceID:          e.EvidenceID,
			Timestamp:           e.Timestamp,
			ContributionContext: e.ContributionContext,
			FileName:            e.FileName,
		}
	}
	WriteJSON(w, http.StatusOK, result)
}

func (a *API) HandleSkillSessions(w http.ResponseWriter, r *http.Request) {
	skillKey := strings.TrimSpace(r.URL.Query().Get("skill_key"))
	if skillKey == "" {
		WriteError(w, http.StatusBadRequest, "skill_key 不能为空")
		return
	}
	if a.rt == nil || a.rt.Core == nil || a.rt.Core.Services.SessionSemantic == nil {
		WriteError(w, http.StatusBadRequest, "会话语义服务未初始化")
		return
	}

	sessions, err := a.rt.Core.Services.SessionSemantic.GetSessionsBySkill(r.Context(), skillKey, 30*24*time.Hour, 10)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]dto.SessionDTO, 0, len(sessions))
	for _, s := range sessions {
		diffIDs := schema.GetInt64Slice(s.Metadata, "diff_ids")
		browserIDs := schema.GetInt64Slice(s.Metadata, "browser_event_ids")
		timeRange := strings.TrimSpace(s.TimeRange)
		if timeRange == "" {
			timeRange = service.FormatTimeRangeMs(s.StartTime, s.EndTime)
		}
		result = append(result, dto.SessionDTO{
			ID:             s.ID,
			Date:           s.Date,
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			TimeRange:      timeRange,
			PrimaryApp:     s.PrimaryApp,
			Category:       s.Category,
			Summary:        s.Summary,
			SkillsInvolved: []string(s.SkillsInvolved),
			DiffCount:      len(diffIDs),
			BrowserCount:   len(browserIDs),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime > result[j].StartTime })
	WriteJSON(w, http.StatusOK, result)
}
