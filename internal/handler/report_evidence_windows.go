//go:build windows

package handler

import (
	"github.com/yuqie6/WorkMirror/internal/dto"
	"github.com/yuqie6/WorkMirror/internal/service"
)

func toSessionRefDTO(r service.EvidenceSessionRef) dto.SessionRefDTO {
	return dto.SessionRefDTO{
		ID:           r.ID,
		Date:         r.Date,
		TimeRange:    r.TimeRange,
		Category:     r.Category,
		Summary:      r.Summary,
		EvidenceHint: r.EvidenceHint,
	}
}

func toEvidenceBlockDTO(sessions []service.EvidenceSessionRef) dto.EvidenceBlockDTO {
	if len(sessions) == 0 {
		return dto.EvidenceBlockDTO{}
	}
	out := make([]dto.SessionRefDTO, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, toSessionRefDTO(s))
	}
	return dto.EvidenceBlockDTO{Sessions: out}
}

func toClaimEvidenceDTO(c service.ClaimEvidence) dto.ClaimEvidenceDTO {
	out := dto.ClaimEvidenceDTO{Claim: c.Claim}
	if len(c.Sessions) == 0 {
		return out
	}
	out.Sessions = make([]dto.SessionRefDTO, 0, len(c.Sessions))
	for _, s := range c.Sessions {
		out.Sessions = append(out.Sessions, toSessionRefDTO(s))
	}
	return out
}

func toDailySummaryEvidenceDTO(ev *service.DailySummaryEvidence) *dto.DailySummaryEvidenceDTO {
	if ev == nil {
		return nil
	}
	out := &dto.DailySummaryEvidenceDTO{
		Summary:    toEvidenceBlockDTO(ev.Summary),
		Highlights: nil,
		Struggles:  nil,
	}
	if len(ev.Highlights) > 0 {
		out.Highlights = make([]dto.ClaimEvidenceDTO, 0, len(ev.Highlights))
		for _, c := range ev.Highlights {
			out.Highlights = append(out.Highlights, toClaimEvidenceDTO(c))
		}
	}
	if len(ev.Struggles) > 0 {
		out.Struggles = make([]dto.ClaimEvidenceDTO, 0, len(ev.Struggles))
		for _, c := range ev.Struggles {
			out.Struggles = append(out.Struggles, toClaimEvidenceDTO(c))
		}
	}
	return out
}

func toPeriodSummaryEvidenceDTO(ev *service.PeriodSummaryEvidence) *dto.PeriodSummaryEvidenceDTO {
	if ev == nil {
		return nil
	}
	out := &dto.PeriodSummaryEvidenceDTO{
		Overview:     toEvidenceBlockDTO(ev.Overview),
		Patterns:     toEvidenceBlockDTO(ev.Patterns),
		Achievements: nil,
		Suggestions:  nil,
	}
	if len(ev.Achievements) > 0 {
		out.Achievements = make([]dto.ClaimEvidenceDTO, 0, len(ev.Achievements))
		for _, c := range ev.Achievements {
			out.Achievements = append(out.Achievements, toClaimEvidenceDTO(c))
		}
	}
	if len(ev.Suggestions) > 0 {
		out.Suggestions = make([]dto.ClaimEvidenceDTO, 0, len(ev.Suggestions))
		for _, c := range ev.Suggestions {
			out.Suggestions = append(out.Suggestions, toClaimEvidenceDTO(c))
		}
	}
	return out
}
