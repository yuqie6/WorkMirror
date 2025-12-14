//go:build windows

package handler

import (
	"net/http"
	"strings"

	"github.com/yuqie6/WorkMirror/internal/dto"
)

func (a *API) HandleDiffDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	id, err := parseInt64Param(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "id 无效")
		return
	}

	if a.rt == nil || a.rt.Repos.Diff == nil {
		WriteError(w, http.StatusBadRequest, "Diff 仓储未初始化")
		return
	}

	diff, err := a.rt.Repos.Diff.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if diff == nil {
		WriteError(w, http.StatusNotFound, "Diff not found")
		return
	}

	var skills []string
	if len(diff.SkillsDetected) > 0 {
		skills = []string(diff.SkillsDetected)
	}

	WriteJSON(w, http.StatusOK, &dto.DiffDetailDTO{
		ID:           diff.ID,
		FileName:     diff.FileName,
		Language:     diff.Language,
		DiffContent:  diff.DiffContent,
		Insight:      diff.AIInsight,
		Skills:       skills,
		LinesAdded:   diff.LinesAdded,
		LinesDeleted: diff.LinesDeleted,
		Timestamp:    diff.Timestamp,
	})
}
