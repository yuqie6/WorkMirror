package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yuqie6/mirror/internal/ai"
	"github.com/yuqie6/mirror/internal/model"
	"github.com/yuqie6/mirror/internal/pkg/config"
	"github.com/yuqie6/mirror/internal/repository"
	"github.com/yuqie6/mirror/internal/service"
)

var (
	cfgFile string
	cfg     *config.Config
	db      *repository.Database
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mirror",
		Short: "Mirror - 智能个人行为量化与成长归因系统",
		Long:  `Mirror 是一个本地运行的 AI 系统，通过自动记录电脑行为，生成学习总结和能力建模。`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 加载配置
			var err error
			cfg, err = config.Load(cfgFile)
			if err != nil {
				slog.Error("加载配置失败", "error", err)
				os.Exit(1)
			}
			config.SetupLogger(cfg.App.LogLevel)

			// 初始化数据库
			db, err = repository.NewDatabase(cfg.Storage.DBPath)
			if err != nil {
				slog.Error("初始化数据库失败", "error", err)
				os.Exit(1)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if db != nil {
				db.Close()
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")

	// 添加子命令
	rootCmd.AddCommand(reportCmd())
	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(statsCmd())
	rootCmd.AddCommand(skillsCmd())
	rootCmd.AddCommand(trendsCmd())
	rootCmd.AddCommand(queryCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// reportCmd 生成报告命令
func reportCmd() *cobra.Command {
	var today bool
	var week bool
	var date string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "生成每日/每周报告",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// 检查 API Key
			if cfg.AI.DeepSeek.APIKey == "" {
				fmt.Println("⚠️  DeepSeek API Key 未配置")
				fmt.Println("   请设置环境变量: DEEPSEEK_API_KEY")
				fmt.Println("   或在 config.yaml 中配置")
				os.Exit(1)
			}

			// 创建服务
			deepseek := ai.NewDeepSeekClient(&ai.DeepSeekConfig{
				APIKey:  cfg.AI.DeepSeek.APIKey,
				BaseURL: cfg.AI.DeepSeek.BaseURL,
				Model:   cfg.AI.DeepSeek.Model,
			})
			analyzer := ai.NewDiffAnalyzer(deepseek)
			diffRepo := repository.NewDiffRepository(db.DB)
			eventRepo := repository.NewEventRepository(db.DB)
			summaryRepo := repository.NewSummaryRepository(db.DB)
			skillRepo := repository.NewSkillRepository(db.DB)
			skillService := service.NewSkillService(skillRepo, diffRepo)
			aiService := service.NewAIService(analyzer, diffRepo, eventRepo, summaryRepo, skillService)

			// 先分析待处理的 Diff
			analyzed, _ := aiService.AnalyzePendingDiffs(ctx, 20)
			if analyzed > 0 {
				fmt.Printf("✅ 已分析 %d 个代码变更\n\n", analyzed)
			}

			if week {
				// 生成周报
				generateWeeklyReport(ctx, aiService, summaryRepo)
			} else {
				// 生成日报
				targetDate := date
				if today || targetDate == "" {
					targetDate = time.Now().Format("2006-01-02")
				}
				generateDailyReport(ctx, aiService, targetDate)
			}
		},
	}

	cmd.Flags().BoolVar(&today, "today", false, "生成今日报告")
	cmd.Flags().BoolVar(&week, "week", false, "生成本周报告")
	cmd.Flags().StringVar(&date, "date", "", "指定日期 (YYYY-MM-DD)")

	return cmd
}

// generateDailyReport 生成日报
func generateDailyReport(ctx context.Context, aiService *service.AIService, targetDate string) {
	fmt.Printf("📊 正在生成 %s 的报告...\n\n", targetDate)

	summary, err := aiService.GenerateDailySummary(ctx, targetDate)
	if err != nil {
		fmt.Printf("❌ 生成报告失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📅 %s 日报\n", targetDate)
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("\n📝 总结\n%s\n", summary.Summary)
	fmt.Printf("\n🌟 亮点\n%s\n", summary.Highlights)
	if summary.Struggles != "" && summary.Struggles != "无" {
		fmt.Printf("\n💪 挑战\n%s\n", summary.Struggles)
	}
	fmt.Printf("\n🎯 技能\n")
	for _, skill := range summary.SkillsGained {
		fmt.Printf("  • %s\n", skill)
	}
	fmt.Printf("\n📊 统计\n")
	fmt.Printf("  • 编码时长: %d 分钟\n", summary.TotalCoding)
	fmt.Printf("  • 代码变更: %d 次\n", summary.TotalDiffs)
	fmt.Println("\n═══════════════════════════════════════")
}

// generateWeeklyReport 生成周报
func generateWeeklyReport(ctx context.Context, aiService *service.AIService, summaryRepo *repository.SummaryRepository) {
	fmt.Println("📊 正在生成本周报告...")

	// 获取最近 7 天的日报
	summaries, err := summaryRepo.GetRecent(ctx, 7)
	if err != nil {
		fmt.Printf("❌ 获取日报失败: %v\n", err)
		os.Exit(1)
	}

	if len(summaries) == 0 {
		fmt.Println("📚 本周还没有日报记录")
		fmt.Println("   先使用 'mirror report --today' 生成日报")
		return
	}

	// 统计汇总
	totalCoding := 0
	totalDiffs := 0
	allSkills := make(map[string]int)

	// 构建请求数据
	dailyInfos := make([]ai.DailySummaryInfo, 0, len(summaries))

	for _, s := range summaries {
		totalCoding += s.TotalCoding
		totalDiffs += s.TotalDiffs
		for _, skill := range s.SkillsGained {
			allSkills[skill]++
		}
		dailyInfos = append(dailyInfos, ai.DailySummaryInfo{
			Date:       s.Date,
			Summary:    s.Summary,
			Highlights: s.Highlights,
			Skills:     []string(s.SkillsGained),
		})
	}

	// 确定日期范围
	startDate := summaries[len(summaries)-1].Date
	endDate := summaries[0].Date

	fmt.Printf("📅 本周周报 (%s ~ %s)\n", startDate, endDate)
	fmt.Println("═══════════════════════════════════════")

	// 调用 AI 生成周报分析
	weeklyResult, err := aiService.GenerateWeeklySummary(ctx, &ai.WeeklySummaryRequest{
		StartDate:      startDate,
		EndDate:        endDate,
		DailySummaries: dailyInfos,
		TotalCoding:    totalCoding,
		TotalDiffs:     totalDiffs,
	})

	if err != nil {
		fmt.Printf("\n⚠️  AI 分析失败: %v\n", err)
		fmt.Println("   显示基础统计信息:")
		// 降级：显示基础统计
		printBasicWeeklyStats(summaries, totalCoding, totalDiffs, allSkills)
		return
	}

	// 输出 AI 分析结果
	fmt.Printf("\n📝 本周概述\n%s\n", weeklyResult.Overview)

	fmt.Printf("\n🏆 主要成就\n")
	for _, a := range weeklyResult.Achievements {
		fmt.Printf("  • %s\n", a)
	}

	fmt.Printf("\n🔍 学习模式\n%s\n", weeklyResult.Patterns)

	fmt.Printf("\n💡 下周建议\n%s\n", weeklyResult.Suggestions)

	fmt.Printf("\n🎯 重点技能\n")
	for _, skill := range weeklyResult.TopSkills {
		fmt.Printf("  • %s\n", skill)
	}

	fmt.Printf("\n📊 本周统计\n")
	fmt.Printf("  • 总编码时长: %d 分钟 (%.1f 小时)\n", totalCoding, float64(totalCoding)/60)
	fmt.Printf("  • 总代码变更: %d 次\n", totalDiffs)
	fmt.Printf("  • 日报天数: %d 天\n", len(summaries))

	fmt.Println("\n═══════════════════════════════════════")
}

// printBasicWeeklyStats 打印基础周统计（AI 失败时的降级方案）
func printBasicWeeklyStats(summaries []model.DailySummary, totalCoding, totalDiffs int, allSkills map[string]int) {
	fmt.Printf("\n📋 每日回顾\n")
	for _, s := range summaries {
		fmt.Printf("  %s: %s\n", s.Date, truncateString(s.Summary, 50))
	}

	fmt.Printf("\n🎯 本周技能 (出现次数)\n")
	for skill, count := range allSkills {
		fmt.Printf("  • %s ×%d\n", skill, count)
	}

	fmt.Printf("\n📊 本周统计\n")
	fmt.Printf("  • 总编码时长: %d 分钟 (%.1f 小时)\n", totalCoding, float64(totalCoding)/60)
	fmt.Printf("  • 总代码变更: %d 次\n", totalDiffs)
	fmt.Printf("  • 日报天数: %d 天\n", len(summaries))

	fmt.Println("\n═══════════════════════════════════════")
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// analyzeCmd 分析命令
func analyzeCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "分析待处理的代码变更",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if cfg.AI.DeepSeek.APIKey == "" {
				fmt.Println("⚠️  DeepSeek API Key 未配置")
				os.Exit(1)
			}

			deepseek := ai.NewDeepSeekClient(&ai.DeepSeekConfig{
				APIKey:  cfg.AI.DeepSeek.APIKey,
				BaseURL: cfg.AI.DeepSeek.BaseURL,
				Model:   cfg.AI.DeepSeek.Model,
			})
			analyzer := ai.NewDiffAnalyzer(deepseek)
			diffRepo := repository.NewDiffRepository(db.DB)
			eventRepo := repository.NewEventRepository(db.DB)
			summaryRepo := repository.NewSummaryRepository(db.DB)
			skillRepo := repository.NewSkillRepository(db.DB)
			skillService := service.NewSkillService(skillRepo, diffRepo)
			aiService := service.NewAIService(analyzer, diffRepo, eventRepo, summaryRepo, skillService)

			fmt.Printf("🔍 正在分析待处理的代码变更 (最多 %d 个)...\n", limit)

			analyzed, err := aiService.AnalyzePendingDiffs(ctx, limit)
			if err != nil {
				fmt.Printf("❌ 分析失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ 已分析 %d 个代码变更\n", analyzed)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "最大分析数量")

	return cmd
}

// statsCmd 统计命令
func statsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "查看统计信息",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			eventRepo := repository.NewEventRepository(db.DB)
			diffRepo := repository.NewDiffRepository(db.DB)

			// 计算时间范围
			now := time.Now()
			endTime := now.UnixMilli()
			startTime := now.AddDate(0, 0, -days).UnixMilli()

			// 事件统计
			eventCount, _ := eventRepo.Count(ctx)
			appStats, _ := eventRepo.GetAppStats(ctx, startTime, endTime)

			// Diff 统计
			diffCount, _ := diffRepo.CountByDateRange(ctx, startTime, endTime)
			langStats, _ := diffRepo.GetLanguageStats(ctx, startTime, endTime)

			fmt.Printf("📊 最近 %d 天统计\n", days)
			fmt.Println("═══════════════════════════════════════")

			fmt.Printf("\n📱 应用使用 (Top 5)\n")
			for i, stat := range appStats {
				if i >= 5 {
					break
				}
				hours := stat.TotalDuration / 3600
				mins := (stat.TotalDuration % 3600) / 60
				fmt.Printf("  • %s: %dh %dm\n", stat.AppName, hours, mins)
			}

			fmt.Printf("\n💻 代码语言 (Top 5)\n")
			for i, stat := range langStats {
				if i >= 5 {
					break
				}
				fmt.Printf("  • %s: %d 次变更, +%d/-%d 行\n",
					stat.Language, stat.DiffCount, stat.LinesAdded, stat.LinesDeleted)
			}

			fmt.Printf("\n📈 总计\n")
			fmt.Printf("  • 窗口事件: %d 条\n", eventCount)
			fmt.Printf("  • 代码变更: %d 次\n", diffCount)
			fmt.Println("\n═══════════════════════════════════════")
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 7, "统计天数")

	return cmd
}

// skillsCmd 技能树命令
func skillsCmd() *cobra.Command {
	var top int

	cmd := &cobra.Command{
		Use:   "skills",
		Short: "查看技能树",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			skillRepo := repository.NewSkillRepository(db.DB)
			diffRepo := repository.NewDiffRepository(db.DB)
			skillService := service.NewSkillService(skillRepo, diffRepo)

			// 获取技能树
			tree, err := skillService.GetSkillTree(ctx)
			if err != nil {
				fmt.Printf("❌ 获取技能树失败: %v\n", err)
				os.Exit(1)
			}

			// 自动修复：如果要展示的技能树为空，但数据库中有已分析的 Diff，则尝试同步
			if tree.TotalSkills == 0 {
				diffs, err := diffRepo.GetAllAnalyzed(ctx)
				if err == nil && len(diffs) > 0 {
					fmt.Printf("🔄 检测到 %d 个已分析的变更但技能树为空，正在同步技能...\n", len(diffs))
					if err := skillService.UpdateSkillsFromDiffs(ctx, diffs); err == nil {
						// 同步后重新获取
						tree, err = skillService.GetSkillTree(ctx)
						if err != nil {
							fmt.Printf("❌ 获取技能树失败: %v\n", err)
							os.Exit(1)
						}
						fmt.Printf("✅ 同步完成，发现 %d 个技能\n\n", tree.TotalSkills)
					} else {
						fmt.Printf("⚠️ 同步技能失败: %v\n", err)
					}
				}
			}

			if tree.TotalSkills == 0 {
				fmt.Println("📚 还没有技能记录")
				fmt.Println("   使用 'mirror analyze' 分析代码变更来积累技能")
				return
			}

			fmt.Printf("🌳 技能树 (共 %d 个技能)\n", tree.TotalSkills)
			fmt.Println("═══════════════════════════════════════")

			// 按分类显示
			categoryNames := map[string]string{
				"language": "💻 编程语言",
				"frontend": "🎨 前端",
				"backend":  "⚙️ 后端",
				"devops":   "🔧 DevOps",
				"data":     "📊 数据",
				"skill":    "🎯 技能",
				"other":    "📦 其他",
			}

			for category, skills := range tree.Categories {
				if len(skills) == 0 {
					continue
				}

				categoryName := categoryNames[category]
				if categoryName == "" {
					categoryName = "📦 " + category
				}

				fmt.Printf("\n%s\n", categoryName)

				count := 0
				for _, skill := range skills {
					if top > 0 && count >= top {
						break
					}

					// 进度条
					barWidth := 20
					filled := int(skill.Progress / 100 * float64(barWidth))
					bar := ""
					for i := 0; i < barWidth; i++ {
						if i < filled {
							bar += "█"
						} else {
							bar += "░"
						}
					}

					trend := ""
					switch skill.Trend {
					case "up":
						trend = "↑"
					case "down":
						trend = "↓"
					default:
						trend = "→"
					}

					fmt.Printf("  %s Lv.%d %s [%s] %.0f%%\n",
						skill.Name, skill.Level, trend, bar, skill.Progress)
					count++
				}
			}

			fmt.Println("\n═══════════════════════════════════════")
		},
	}

	cmd.Flags().IntVarP(&top, "top", "n", 0, "每个分类显示前 N 个技能 (0=全部)")

	return cmd
}

// trendsCmd 趋势分析命令
func trendsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "查看技能和编码趋势",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			skillRepo := repository.NewSkillRepository(db.DB)
			diffRepo := repository.NewDiffRepository(db.DB)
			eventRepo := repository.NewEventRepository(db.DB)
			trendService := service.NewTrendService(skillRepo, diffRepo, eventRepo)

			period := service.TrendPeriod7Days
			if days == 30 {
				period = service.TrendPeriod30Days
			}

			report, err := trendService.GetTrendReport(ctx, period)
			if err != nil {
				fmt.Printf("❌ 获取趋势失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("📈 趋势分析 (%s - %s)\n", report.StartDate, report.EndDate)
			fmt.Println("═══════════════════════════════════════")

			// 语言分布
			fmt.Printf("\n💻 编程语言分布\n")
			for _, lang := range report.TopLanguages {
				bar := ""
				width := int(lang.Percentage / 5)
				for i := 0; i < width; i++ {
					bar += "█"
				}
				fmt.Printf("  %s: %s %.1f%% (%d次)\n", lang.Language, bar, lang.Percentage, lang.DiffCount)
			}

			// 技能状态
			fmt.Printf("\n🎯 技能状态\n")
			for _, skill := range report.TopSkills {
				status := ""
				switch skill.Status {
				case "growing":
					status = "🔼"
				case "declining":
					status = "🔽"
				default:
					status = "➡️"
				}
				fmt.Printf("  %s %s (%d天活跃)\n", status, skill.SkillName, skill.DaysActive)
			}

			// 统计
			fmt.Printf("\n📊 期间统计\n")
			fmt.Printf("  • 代码变更: %d 次 (日均 %.1f)\n", report.TotalDiffs, report.AvgDiffsPerDay)
			fmt.Printf("  • 编码时长: %d 分钟 (%.1f 小时)\n", report.TotalCodingMins, float64(report.TotalCodingMins)/60)

			// 瓶颈
			if len(report.Bottlenecks) > 0 {
				fmt.Printf("\n⚠️ 需要关注\n")
				for _, b := range report.Bottlenecks {
					fmt.Printf("  • %s\n", b)
				}
			}

			fmt.Println("\n═══════════════════════════════════════")
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 7, "分析天数 (7 或 30)")

	return cmd
}

// queryCmd 查询历史记忆
func queryCmd() *cobra.Command {
	var topK int

	cmd := &cobra.Command{
		Use:   "query [问题]",
		Short: "查询历史学习记忆 (RAG)",
		Long:  "使用语义搜索查询历史编程活动和学习记录",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.Join(args, " ")

			cfg, err := config.Load(cfgFile)
			if err != nil {
				fmt.Printf("❌ 加载配置失败: %v\n", err)
				return
			}

			// 初始化数据库
			db, err := repository.NewDatabase(cfg.Storage.DBPath)
			if err != nil {
				fmt.Printf("❌ 初始化数据库失败: %v\n", err)
				return
			}
			defer db.Close()

			// 创建仓储
			summaryRepo := repository.NewSummaryRepository(db.DB)
			diffRepo := repository.NewDiffRepository(db.DB)

			// 创建 SiliconFlow 客户端
			sfClient := ai.NewSiliconFlowClient(&ai.SiliconFlowConfig{
				APIKey:         cfg.AI.SiliconFlow.APIKey,
				BaseURL:        cfg.AI.SiliconFlow.BaseURL,
				EmbeddingModel: cfg.AI.SiliconFlow.EmbeddingModel,
				RerankerModel:  cfg.AI.SiliconFlow.RerankerModel,
			})

			if !sfClient.IsConfigured() {
				fmt.Println("❌ SiliconFlow API 未配置，无法使用 RAG 查询")
				fmt.Println("请在 config.yaml 中配置 ai.siliconflow.api_key")
				return
			}

			// 创建 RAG 服务
			ragService, err := service.NewRAGService(sfClient, summaryRepo, diffRepo, nil)
			if err != nil {
				fmt.Printf("❌ 初始化 RAG 服务失败: %v\n", err)
				return
			}
			defer ragService.Close()

			ctx := context.Background()

			fmt.Printf("\n🔍 搜索: %s\n\n", query)

			results, err := ragService.Query(ctx, query, topK)
			if err != nil {
				fmt.Printf("❌ 查询失败: %v\n", err)
				return
			}

			if len(results) == 0 {
				fmt.Println("未找到相关记忆，请先运行 mirror analyze 分析代码并生成总结")
				return
			}

			fmt.Printf("📚 找到 %d 条相关记忆:\n\n", len(results))
			for i, r := range results {
				fmt.Printf("──────────────────────────────────────\n")
				fmt.Printf("[%d] 类型: %s | 日期: %s | 相似度: %.2f\n", i+1, r.Type, r.Date, r.Similarity)
				fmt.Printf("%s\n", r.Content)
			}
			fmt.Println("──────────────────────────────────────")
		},
	}

	cmd.Flags().IntVarP(&topK, "top", "n", 5, "返回结果数量")

	return cmd
}
