//go:build windows

package observability

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/yuqie6/mirror/internal/bootstrap"
	"github.com/yuqie6/mirror/internal/dto"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/pkg/privacy"
)

var ErrNotReady = errors.New("rt not ready")

func WriteDiagnosticsZip(ctx context.Context, w io.Writer, rt *bootstrap.AgentRuntime, startedAt time.Time) error {
	if rt == nil || rt.Cfg == nil {
		return ErrNotReady
	}

	status, err := BuildStatus(ctx, rt, startedAt)
	if err != nil {
		return err
	}

	return WriteDiagnosticsZipWithStatus(w, rt, status)
}

func WriteDiagnosticsZipWithStatus(w io.Writer, rt *bootstrap.AgentRuntime, status *dto.StatusDTO) error {
	if rt == nil || rt.Cfg == nil {
		return ErrNotReady
	}
	if status == nil {
		return errors.New("status is nil")
	}

	sanitizer := privacy.New(rt.Cfg.Privacy.Enabled, rt.Cfg.Privacy.Patterns)

	zw := zip.NewWriter(w)
	defer zw.Close()

	_ = addZipJSON(zw, "status.json", status)
	_ = addZipText(zw, "README.txt", buildDiagReadme())

	cfgPath, _ := config.DefaultConfigPath()
	if strings.TrimSpace(cfgPath) != "" {
		if b, err := os.ReadFile(cfgPath); err == nil {
			_ = addZipText(zw, "config/config.yaml.redacted", redactConfigYAML(string(b)))
		} else {
			_ = addZipText(zw, "config/ERROR.txt", "读取配置失败: "+err.Error())
		}
	}

	logPath := strings.TrimSpace(rt.Cfg.App.LogPath)
	if logPath != "" {
		lines, err := tailLines(logPath, 512*1024)
		if err != nil {
			_ = addZipText(zw, "logs/ERROR.txt", "读取日志失败: "+err.Error())
		} else {
			if len(lines) > 2000 {
				lines = lines[len(lines)-2000:]
			}
			for i := range lines {
				lines[i] = sanitizeLogLine(lines[i], sanitizer)
			}
			_ = addZipText(zw, "logs/recent.log", strings.Join(lines, "\n"))
		}
	}

	return nil
}

func ReadRecentErrors(logPath string, sanitizer *privacy.Sanitizer, limit int) []dto.RecentErrorDTO {
	path := strings.TrimSpace(logPath)
	if path == "" {
		return nil
	}
	lines, err := tailLines(path, 256*1024)
	if err != nil {
		return []dto.RecentErrorDTO{{Message: "读取日志失败: " + err.Error()}}
	}

	if limit <= 0 {
		limit = 20
	}

	out := make([]dto.RecentErrorDTO, 0, limit)
	for i := len(lines) - 1; i >= 0 && len(out) < limit; i-- {
		raw := strings.TrimSpace(lines[i])
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "level=ERROR") &&
			!strings.Contains(raw, "level=WARN") &&
			!strings.Contains(raw, "level=error") &&
			!strings.Contains(raw, "level=warn") {
			continue
		}
		redacted := sanitizeLogLine(raw, sanitizer)
		out = append(out, parseLogLine(redacted))
	}
	return out
}

func buildDiagReadme() string {
	return strings.TrimSpace(`
该诊断包默认不包含数据库与原始证据数据。

包含：
- status.json：/api/status 快照
- config/config.yaml.redacted：脱敏后的配置文件（如存在）
- logs/recent.log：最近日志（截断 + 脱敏）

建议在提交 issue/反馈时附上该文件。`) + "\n"
}

func addZipText(zw *zip.Writer, name string, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func addZipJSON(zw *zip.Writer, name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return addZipText(zw, name, string(b)+"\n")
}

var reURL = regexp.MustCompile(`https?://[^\s]+`)

func sanitizeLogLine(line string, sanitizer *privacy.Sanitizer) string {
	s := strings.TrimRight(line, "\r\n")
	if sanitizer == nil || !sanitizer.Enabled() {
		return s
	}

	s = reURL.ReplaceAllStringFunc(s, func(m string) string {
		return sanitizer.SanitizeURL(m)
	})
	return sanitizer.SanitizeText(s)
}

var reYAMLSecret = regexp.MustCompile(`(?im)^(\s*(api_key|apikey|access_token|refresh_token|token|secret)\s*:\s*)(.+)$`)

func redactConfigYAML(y string) string {
	in := strings.ReplaceAll(y, "\r\n", "\n")
	return reYAMLSecret.ReplaceAllStringFunc(in, func(line string) string {
		m := reYAMLSecret.FindStringSubmatch(line)
		if len(m) != 4 {
			return line
		}
		prefix := m[1]
		val := strings.TrimSpace(m[3])
		if strings.Contains(val, "${") {
			return prefix + val
		}
		return prefix + "\"***\""
	})
}

func tailLines(path string, maxBytes int64) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := st.Size()
	start := int64(0)
	if size > maxBytes {
		start = size - maxBytes
	}
	if start > 0 {
		if _, err := f.Seek(start, io.SeekStart); err != nil {
			return nil, err
		}
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	s := string(b)
	if start > 0 {
		if idx := strings.IndexByte(s, '\n'); idx >= 0 {
			s = s[idx+1:]
		}
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n"), nil
}
