package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	chromem "github.com/philippgille/chromem-go"
	"github.com/yuqie6/WorkMirror/internal/ai"
	"github.com/yuqie6/WorkMirror/internal/repository"
	"github.com/yuqie6/WorkMirror/internal/schema"
)

// RAGService RAG 长期记忆服务
type RAGService struct {
	db          *chromem.DB
	collection  *chromem.Collection
	sfClient    *ai.SiliconFlowClient
	summaryRepo *repository.SummaryRepository
	diffRepo    *repository.DiffRepository
	storagePath string
}

// RAGConfig 配置
type RAGConfig struct {
	StoragePath string // 向量数据库存储路径
}

// NewRAGService 创建 RAG 服务
func NewRAGService(
	sfClient *ai.SiliconFlowClient,
	summaryRepo *repository.SummaryRepository,
	diffRepo *repository.DiffRepository,
	cfg *RAGConfig,
) (*RAGService, error) {
	if cfg == nil {
		cfg = &RAGConfig{}
	}

	if cfg.StoragePath == "" {
		cfg.StoragePath = "./data/rag"
	}

	// 确保目录存在
	if err := os.MkdirAll(cfg.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("创建 RAG 存储目录失败: %w", err)
	}

	// 创建向量数据库
	db, err := chromem.NewPersistentDB(cfg.StoragePath, false)
	if err != nil {
		return nil, fmt.Errorf("创建向量数据库失败: %w", err)
	}

	// 创建或获取 collection
	collection, err := db.GetOrCreateCollection("memories", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 collection 失败: %w", err)
	}

	return &RAGService{
		db:          db,
		collection:  collection,
		sfClient:    sfClient,
		summaryRepo: summaryRepo,
		diffRepo:    diffRepo,
		storagePath: cfg.StoragePath,
	}, nil
}

// IndexDailySummary 索引每日总结
func (s *RAGService) IndexDailySummary(ctx context.Context, summary *schema.DailySummary) error {
	if !s.sfClient.IsConfigured() {
		slog.Debug("SiliconFlow 未配置，跳过索引")
		return nil
	}

	// 构建文档内容
	content := fmt.Sprintf("日期: %s\n总结: %s\n亮点: %s\n挑战: %s",
		summary.Date, summary.Summary, summary.Highlights, summary.Struggles)

	// 生成嵌入
	embeddings, err := s.sfClient.Embed(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %w", err)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("嵌入结果为空")
	}

	// 添加到向量数据库
	doc := chromem.Document{
		ID:        fmt.Sprintf("summary_%s", summary.Date),
		Content:   content,
		Embedding: embeddings[0],
		Metadata: map[string]string{
			"type": "daily_summary",
			"date": summary.Date,
		},
	}

	if err := s.collection.AddDocument(ctx, doc); err != nil {
		return fmt.Errorf("添加文档失败: %w", err)
	}

	slog.Debug("索引每日总结", "date", summary.Date)
	return nil
}

// IndexDiff 索引代码变更
func (s *RAGService) IndexDiff(ctx context.Context, diff *schema.Diff) error {
	if !s.sfClient.IsConfigured() {
		return nil
	}

	// 只索引有 AI 解读的 diff
	if diff.AIInsight == "" {
		return nil
	}

	content := fmt.Sprintf("文件: %s\n语言: %s\n解读: %s",
		diff.FileName, diff.Language, diff.AIInsight)

	embeddings, err := s.sfClient.Embed(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %w", err)
	}

	if len(embeddings) == 0 {
		return nil
	}

	doc := chromem.Document{
		ID:        fmt.Sprintf("diff_%d", diff.ID),
		Content:   content,
		Embedding: embeddings[0],
		Metadata: map[string]string{
			"type":     "diff",
			"file":     diff.FileName,
			"language": diff.Language,
			"date":     time.UnixMilli(diff.Timestamp).Format("2006-01-02"),
		},
	}

	if err := s.collection.AddDocument(ctx, doc); err != nil {
		return fmt.Errorf("添加文档失败: %w", err)
	}

	slog.Debug("索引 Diff", "file", diff.FileName)
	return nil
}

// Query 查询相关记忆
func (s *RAGService) Query(ctx context.Context, query string, topK int) ([]MemoryResult, error) {
	if !s.sfClient.IsConfigured() {
		return nil, fmt.Errorf("SiliconFlow 未配置")
	}

	if topK <= 0 {
		topK = 5
	}

	// 生成查询嵌入
	queryEmb, err := s.sfClient.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("生成查询嵌入失败: %w", err)
	}

	if len(queryEmb) == 0 {
		return nil, fmt.Errorf("查询嵌入为空")
	}

	// 向量搜索 (使用余弦相似度)
	results, err := s.collection.QueryEmbedding(ctx, queryEmb[0], topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 提取文档内容用于重排
	docs := make([]string, len(results))
	for i, r := range results {
		docs[i] = r.Content
	}

	// 使用 Reranker 重排
	reranked, err := s.sfClient.Rerank(ctx, query, docs, topK)
	if err != nil {
		slog.Warn("重排失败，使用原始结果", "error", err)
		// 返回原始结果
		memories := make([]MemoryResult, len(results))
		for i, r := range results {
			memories[i] = MemoryResult{
				Content:    r.Content,
				Similarity: r.Similarity,
				Type:       r.Metadata["type"],
				Date:       r.Metadata["date"],
			}
		}
		return memories, nil
	}

	// 按重排结果返回
	memories := make([]MemoryResult, len(reranked))
	for i, rr := range reranked {
		if rr.Index < len(results) {
			memories[i] = MemoryResult{
				Content:    results[rr.Index].Content,
				Similarity: float32(rr.RelevanceScore),
				Type:       results[rr.Index].Metadata["type"],
				Date:       results[rr.Index].Metadata["date"],
			}
		}
	}

	return memories, nil
}

// MemoryResult 记忆查询结果
type MemoryResult struct {
	Content    string
	Similarity float32
	Type       string
	Date       string
}

// Close 关闭服务
func (s *RAGService) Close() error {
	// chromem-go 持久化数据库会自动保存
	return nil
}

// GetStoragePath 获取存储路径
func (s *RAGService) GetStoragePath() string {
	absPath, _ := filepath.Abs(s.storagePath)
	return absPath
}
