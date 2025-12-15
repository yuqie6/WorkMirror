//go:build windows

package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuqie6/WorkMirror/internal/bootstrap"
	"github.com/yuqie6/WorkMirror/internal/eventbus"
	"github.com/yuqie6/WorkMirror/internal/handler"
	"github.com/yuqie6/WorkMirror/internal/uiassets"
)

// LocalServer 本地 HTTP 服务器
type LocalServer struct {
	rt      *bootstrap.AgentRuntime
	hub     *eventbus.Hub
	ln      net.Listener
	srv     *http.Server
	baseURL string
}

// Options 服务器启动配置
type Options struct {
	ListenAddr string // e.g. "127.0.0.1:0"
}

// Start 启动本地 HTTP 服务器
func Start(ctx context.Context, rt *bootstrap.AgentRuntime, opts Options) (*LocalServer, error) {
	if rt == nil {
		return nil, fmt.Errorf("rt 不能为空")
	}
	if strings.TrimSpace(opts.ListenAddr) == "" {
		opts.ListenAddr = "127.0.0.1:0"
	}

	ln, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return nil, err
	}

	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		_ = ln.Close()
		return nil, err
	}
	baseURL := "http://127.0.0.1:" + portStr

	hub := rt.Hub
	if hub == nil {
		hub = eventbus.NewHub()
	}

	api := handler.NewAPI(rt, hub)

	mux := http.NewServeMux()
	registerRoutes(mux, api)

	uiFS, uiSource := pickUIFS()
	mux.Handle("/", spaHandler(uiFS, "index.html"))
	slog.Info("UI 资源来源", "source", uiSource)

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ls := &LocalServer{
		rt:      rt,
		hub:     hub,
		ln:      ln,
		srv:     srv,
		baseURL: baseURL,
	}

	go func() {
		<-ctx.Done()
		_ = ls.Shutdown(context.Background())
	}()

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server 异常退出", "error", err)
		}
	}()

	writeBaseURLFile(baseURL)
	slog.Info("本地 HTTP 已启动", "base_url", baseURL)
	return ls, nil
}

// registerRoutes 注册所有 API 路由
func registerRoutes(mux *http.ServeMux, api *handler.API) {
	mux.HandleFunc("/health", api.HandleHealth)
	mux.HandleFunc("/api/events", api.HandleSSE)
	mux.HandleFunc("/api/status", requireMethod(http.MethodGet, api.HandleStatus))

	mux.HandleFunc("/api/summary/today", requireMethod(http.MethodGet, api.HandleTodaySummary))
	mux.HandleFunc("/api/summary/daily", requireMethod(http.MethodGet, api.HandleDailySummary))
	mux.HandleFunc("/api/summary/index", requireMethod(http.MethodGet, api.HandleSummaryIndex))
	mux.HandleFunc("/api/summary/period", requireMethod(http.MethodGet, api.HandlePeriodSummary))
	mux.HandleFunc("/api/summary/period/index", requireMethod(http.MethodGet, api.HandlePeriodSummaryIndex))

	mux.HandleFunc("/api/skills/tree", requireMethod(http.MethodGet, api.HandleSkillTree))
	mux.HandleFunc("/api/skills/evidence", requireMethod(http.MethodGet, api.HandleSkillEvidence))
	mux.HandleFunc("/api/skills/sessions", requireMethod(http.MethodGet, api.HandleSkillSessions))

	mux.HandleFunc("/api/trends", requireMethod(http.MethodGet, api.HandleTrends))
	mux.HandleFunc("/api/app-stats", requireMethod(http.MethodGet, api.HandleAppStats))

	mux.HandleFunc("/api/diffs/detail", requireMethod(http.MethodGet, api.HandleDiffDetail))

	mux.HandleFunc("/api/sessions/by-date", requireMethod(http.MethodGet, api.HandleSessionsByDate))
	mux.HandleFunc("/api/sessions/detail", requireMethod(http.MethodGet, api.HandleSessionDetail))
	mux.HandleFunc("/api/sessions/events", requireMethod(http.MethodGet, api.HandleSessionEvents))
	mux.HandleFunc("/api/sessions/build", requireMethod(http.MethodPost, api.HandleBuildSessionsForDate))
	mux.HandleFunc("/api/sessions/rebuild", requireMethod(http.MethodPost, api.HandleRebuildSessionsForDate))
	mux.HandleFunc("/api/sessions/enrich", requireMethod(http.MethodPost, api.HandleEnrichSessionsForDate))

	// v0.2 product API aliases
	mux.HandleFunc("/api/maintenance/sessions/rebuild", requireMethod(http.MethodPost, api.HandleMaintenanceSessionsRebuild))
	mux.HandleFunc("/api/maintenance/sessions/enrich", requireMethod(http.MethodPost, api.HandleMaintenanceSessionsEnrich))
	mux.HandleFunc("/api/maintenance/sessions/repair-evidence", requireMethod(http.MethodPost, api.HandleMaintenanceSessionsRepairEvidence))

	mux.HandleFunc("/api/diagnostics/export", requireMethod(http.MethodGet, api.HandleDiagnosticsExport))

	mux.HandleFunc("/api/settings", api.HandleSettings)
}

// requireMethod 创建要求特定 HTTP 方法的中间件
func requireMethod(method string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		fn(w, r)
	}
}

// BaseURL 返回服务器的基础 URL
func (s *LocalServer) BaseURL() string {
	if s == nil {
		return ""
	}
	return s.baseURL
}

// Shutdown 优雅关闭服务器
func (s *LocalServer) Shutdown(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

// writeBaseURLFile 将服务地址写入文件供外部读取
func writeBaseURLFile(baseURL string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exe)
	dataDir := filepath.Join(exeDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dataDir, "http_base_url.txt"), []byte(baseURL), 0o644)
}

// spaHandler 创建 SPA 静态资源处理器
// 支持前端路由回退到 index.html
func spaHandler(assetFS fs.FS, index string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if upath == "" || upath == "/" {
			serveAsset(w, r, assetFS, index)
			return
		}

		clean := path.Clean(upath)
		clean = strings.TrimPrefix(clean, "/")

		// 只允许访问 dist 内文件
		if strings.Contains(clean, "..") {
			http.NotFound(w, r)
			return
		}

		// 资源存在则直接返回，否则回退到 index.html（SPA 路由）
		if _, err := fs.Stat(assetFS, clean); err == nil {
			serveAsset(w, r, assetFS, clean)
			return
		}
		serveAsset(w, r, assetFS, index)
	})
}

// pickUIFS 选择 UI 静态资源来源
// 优先使用本地 frontend/dist 目录，否则回退到内嵌资源
func pickUIFS() (fs.FS, string) {
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(exeDir, "frontend", "dist"),
		}
		for _, dir := range candidates {
			if dir == "" {
				continue
			}
			indexPath := filepath.Join(dir, "index.html")
			if _, err := os.Stat(indexPath); err != nil {
				continue
			}
			if b, err := os.ReadFile(indexPath); err == nil {
				// 避免误用旧的 Wails 构建产物（会引用 wailsjs，导致纯 HTTP 模式不可用）
				if bytes.Contains(b, []byte("wailsjs")) {
					continue
				}
			}
			return os.DirFS(dir), dir
		}
	}
	return uiassets.FS(), "embedded"
}

// serveAsset 提供静态资源文件
func serveAsset(w http.ResponseWriter, r *http.Request, assetFS fs.FS, name string) {
	f, err := assetFS.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	if ctype := mime.TypeByExtension(path.Ext(name)); ctype != "" {
		w.Header().Set("Content-Type", ctype)
	}
	payload, err := io.ReadAll(f)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, name, stat.ModTime(), bytes.NewReader(payload))
}
