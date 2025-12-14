//go:build windows

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/yuqie6/mirror/internal/observability"
)

func (a *API) HandleDiagnosticsExport(w http.ResponseWriter, r *http.Request) {
	if a == nil || a.rt == nil || a.rt.Cfg == nil {
		WriteAPIError(w, http.StatusServiceUnavailable, APIError{
			Error: "rt 未初始化",
			Code:  "rt_not_ready",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	status, err := observability.BuildStatus(ctx, a.rt, a.startTime)
	if err != nil {
		WriteAPIError(w, http.StatusInternalServerError, APIError{
			Error: err.Error(),
			Code:  "status_build_failed",
		})
		return
	}

	filename := "mirror-diagnostics-" + time.Now().Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	_ = observability.WriteDiagnosticsZipWithStatus(w, a.rt, status)
}
