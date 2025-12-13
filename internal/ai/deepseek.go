package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// DeepSeekClient DeepSeek API 客户端
type DeepSeekClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// DeepSeekConfig 配置
type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewDeepSeekClient 创建客户端
func NewDeepSeekClient(cfg *DeepSeekConfig) *DeepSeekClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}

	return &DeepSeekClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat 发送聊天请求
func (c *DeepSeekClient) Chat(ctx context.Context, messages []Message) (string, error) {
	return c.ChatWithOptions(ctx, messages, 0.3, 2000)
}

// ChatWithOptions 带参数的聊天请求
func (c *DeepSeekClient) ChatWithOptions(ctx context.Context, messages []Message, temperature float64, maxTokens int) (string, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("DeepSeek API 错误", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("API 错误: %s", resp.Status)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("无响应内容")
	}

	slog.Debug("DeepSeek API 调用成功",
		"tokens", chatResp.Usage.TotalTokens,
		"model", c.model,
	)

	return chatResp.Choices[0].Message.Content, nil
}

// IsConfigured 检查是否已配置
func (c *DeepSeekClient) IsConfigured() bool {
	return c.apiKey != ""
}

// ChatWithRetry 带重试的聊天请求（指数退避）
func (c *DeepSeekClient) ChatWithRetry(ctx context.Context, messages []Message, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := c.Chat(ctx, messages)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// 检查是否是可重试错误（网络错误、5xx 错误）
		if !isRetryableError(err) {
			return "", err
		}

		// 指数退避：1s, 2s, 4s...
		backoff := time.Duration(1<<uint(i)) * time.Second
		slog.Warn("API 调用失败，准备重试", "attempt", i+1, "backoff", backoff, "error", err)

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", fmt.Errorf("达到最大重试次数 (%d): %w", maxRetries, lastErr)
}

// isRetryableError 判断是否是可重试错误
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// 网络错误或 5xx 错误可重试
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504")
}
