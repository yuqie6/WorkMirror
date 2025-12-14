//go:build windows

package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type APIError struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
	Hint  string `json:"hint,omitempty"`
}

// WriteJSON 将数据序列化为 JSON 并写入响应
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteAPIError(w http.ResponseWriter, status int, e APIError) {
	if strings.TrimSpace(e.Error) == "" {
		e.Error = http.StatusText(status)
	}
	WriteJSON(w, status, e)
}

// WriteError 写入错误响应
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteAPIError(w, status, APIError{Error: msg})
}

// readJSON 从请求体读取并解析 JSON
func readJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

// parseInt64Param 解析字符串为 int64 参数
func parseInt64Param(value string) (int64, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, fmt.Errorf("参数为空")
	}
	return strconv.ParseInt(v, 10, 64)
}

// strconvAtoiSafe 安全地将字符串转换为整数
func strconvAtoiSafe(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}
