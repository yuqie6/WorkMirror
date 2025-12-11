package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/danielsclee/mirror/internal/ai"
	"github.com/danielsclee/mirror/internal/pkg/config"
	"github.com/danielsclee/mirror/internal/repository"
	"github.com/danielsclee/mirror/internal/service"
	"github.com/spf13/cobra"
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// reportCmd 生成报告命令
func reportCmd() *cobra.Command {
	var today bool
	var date string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "生成每日/每周报告",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// 确定日期
			targetDate := date
			if today || targetDate == "" {
				targetDate = time.Now().Format("2006-01-02")
			}

			fmt.Printf("📊 正在生成 %s 的报告...\n\n", targetDate)

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
			aiService := service.NewAIService(analyzer, diffRepo, eventRepo, summaryRepo)

			// 先分析待处理的 Diff
			analyzed, _ := aiService.AnalyzePendingDiffs(ctx, 20)
			if analyzed > 0 {
				fmt.Printf("✅ 已分析 %d 个代码变更\n\n", analyzed)
			}

			// 生成每日总结
			summary, err := aiService.GenerateDailySummary(ctx, targetDate)
			if err != nil {
				fmt.Printf("❌ 生成报告失败: %v\n", err)
				os.Exit(1)
			}

			// 输出报告
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
		},
	}

	cmd.Flags().BoolVar(&today, "today", false, "生成今日报告")
	cmd.Flags().StringVar(&date, "date", "", "指定日期 (YYYY-MM-DD)")

	return cmd
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
			aiService := service.NewAIService(analyzer, diffRepo, eventRepo, summaryRepo)

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
