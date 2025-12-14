//go:build windows

package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/eventbus"
)

// API HTTP 处理器
type API struct {
	rt        *bootstrap.AgentRuntime
	hub       *eventbus.Hub
	startTime time.Time
}

// NewAPI 创建 API 处理器
func NewAPI(rt *bootstrap.AgentRuntime, hub *eventbus.Hub) *API {
	return &API{
		rt:        rt,
		hub:       hub,
		startTime: time.Now(),
	}
}

// HandleHealth 健康检查接口
func (a *API) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if a == nil || a.rt == nil || a.rt.Cfg == nil {
		WriteError(w, http.StatusServiceUnavailable, "rt 未初始化")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"name":       a.rt.Cfg.App.Name,
		"version":    a.rt.Cfg.App.Version,
		"started_at": a.startTime.Format(time.RFC3339),
	})
}

// HandleSSE Server-Sent Events 接口
func (a *API) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "stream not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if a == nil || a.hub == nil {
		WriteError(w, http.StatusServiceUnavailable, "hub 未初始化")
		return
	}

	ctx := r.Context()
	sub := a.hub.Subscribe(ctx, 32)

	_, _ = io.WriteString(w, "event: ready\n")
	_, _ = io.WriteString(w, "data: {}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = io.WriteString(w, "event: ping\n")
			_, _ = io.WriteString(w, "data: {}\n\n")
			flusher.Flush()
		case evt, ok := <-sub:
			if !ok {
				return
			}
			b, _ := json.Marshal(evt)
			_, _ = io.WriteString(w, "event: "+sanitizeSSEName(evt.Type)+"\n")
			_, _ = io.WriteString(w, "data: ")
			_, _ = w.Write(b)
			_, _ = io.WriteString(w, "\n\n")
			flusher.Flush()
		}
	}
}

// sanitizeSSEName 清理 SSE 事件名称
func sanitizeSSEName(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return "message"
	}
	n = strings.ReplaceAll(n, "\n", "")
	n = strings.ReplaceAll(n, "\r", "")
	return n
}

// Subscribe 订阅事件总线
func (a *API) Subscribe(ctx context.Context, buffer int) <-chan eventbus.Event {
	if a == nil || a.hub == nil {
		ch := make(chan eventbus.Event)
		close(ch)
		return ch
	}
	return a.hub.Subscribe(ctx, buffer)
}

func (a *API) requireWritableDB(w http.ResponseWriter) bool {
	if a == nil || a.rt == nil || a.rt.Core == nil || a.rt.Core.DB == nil {
		WriteAPIError(w, http.StatusServiceUnavailable, APIError{
			Error: "数据库未初始化",
			Code:  "db_not_ready",
			Hint:  "请稍后重试；若持续失败，请导出诊断包并检查日志",
		})
		return false
	}
	if a.rt.Core.DB.SafeMode {
		WriteAPIError(w, http.StatusServiceUnavailable, APIError{
			Error: "数据库处于安全模式，已禁用写入操作",
			Code:  "db_safe_mode",
			Hint:  "请先在 Status 页查看原因并导出诊断包；修复后重启 Agent",
		})
		return false
	}
	return true
}
