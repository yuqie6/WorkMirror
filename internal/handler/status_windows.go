//go:build windows

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/yuqie6/mirror/internal/observability"
)

func (a *API) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if a == nil || a.rt == nil || a.rt.Cfg == nil || a.rt.Core == nil || a.rt.Core.DB == nil {
		WriteAPIError(w, http.StatusServiceUnavailable, APIError{
			Error: "rt 未初始化",
			Code:  "rt_not_ready",
			Hint:  "请稍后重试；若持续失败，请查看日志或重新启动 Agent",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	st, err := observability.BuildStatus(ctx, a.rt, a.startTime)
	if err != nil {
		WriteAPIError(w, http.StatusInternalServerError, APIError{
			Error: err.Error(),
			Code:  "status_build_failed",
			Hint:  "请查看 /api/diagnostics/export 导出的诊断包（或检查日志）",
		})
		return
	}
	WriteJSON(w, http.StatusOK, st)
}
