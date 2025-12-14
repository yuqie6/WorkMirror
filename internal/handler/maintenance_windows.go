//go:build windows

package handler

import "net/http"

// v0.2：为产品化 API 口径提供别名路由（保持历史兼容）。

func (a *API) HandleMaintenanceSessionsRebuild(w http.ResponseWriter, r *http.Request) {
	a.HandleRebuildSessionsForDate(w, r)
}

func (a *API) HandleMaintenanceSessionsEnrich(w http.ResponseWriter, r *http.Request) {
	a.HandleEnrichSessionsForDate(w, r)
}
