//go:build windows

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
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
	"strconv"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/eventbus"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/uiassets"
)

type LocalServer struct {
	rt      *bootstrap.AgentRuntime
	hub     *eventbus.Hub
	ln      net.Listener
	srv     *http.Server
	baseURL string
}

type Options struct {
	ListenAddr string // e.g. "127.0.0.1:0"
}

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

	api := newAPI(rt, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.handleHealth)
	mux.HandleFunc("/api/events", api.handleSSE)
	api.registerJSONRoutes(mux)

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

func (s *LocalServer) BaseURL() string {
	if s == nil {
		return ""
	}
	return s.baseURL
}

func (s *LocalServer) Shutdown(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

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

type apiServer struct {
	rt        *bootstrap.AgentRuntime
	hub       *eventbus.Hub
	cfgPath   string
	startTime time.Time
}

func newAPI(rt *bootstrap.AgentRuntime, hub *eventbus.Hub) *apiServer {
	cfgPath, _ := config.DefaultConfigPath()
	return &apiServer{
		rt:        rt,
		hub:       hub,
		cfgPath:   cfgPath,
		startTime: time.Now(),
	}
}

func (a *apiServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"name":       a.rt.Cfg.App.Name,
		"version":    a.rt.Cfg.App.Version,
		"started_at": a.startTime.Format(time.RFC3339),
	})
}

func (a *apiServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "stream not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	sub := a.hub.Subscribe(ctx, 32)

	// initial event
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

func sanitizeSSEName(name string) string {
	n := strings.TrimSpace(name)
	if n == "" {
		return "message"
	}
	n = strings.ReplaceAll(n, "\n", "")
	n = strings.ReplaceAll(n, "\r", "")
	return n
}

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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func readJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func parseInt64Param(value string) (int64, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, fmt.Errorf("参数为空")
	}
	return strconv.ParseInt(v, 10, 64)
}
