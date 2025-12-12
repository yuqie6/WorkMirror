package repository

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动
	"github.com/yuqie6/mirror/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(dbPath string) (*Database, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置 SQLite WAL 模式
	if err := configureDB(db); err != nil {
		return nil, fmt.Errorf("配置数据库失败: %w", err)
	}

	// 自动迁移表结构
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("迁移数据库失败: %w", err)
	}

	slog.Info("数据库初始化成功", "path", dbPath)

	return &Database{DB: db}, nil
}

// configureDB 配置 SQLite 性能参数
func configureDB(db *gorm.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",    // 启用 WAL 模式，支持并发读写
		"PRAGMA synchronous=NORMAL",  // 平衡性能与安全
		"PRAGMA cache_size=10000",    // 增加缓存 (~40MB)
		"PRAGMA temp_store=MEMORY",   // 临时表使用内存
		"PRAGMA mmap_size=268435456", // 启用内存映射 (256MB)
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return fmt.Errorf("执行 %s 失败: %w", pragma, err)
		}
	}

	return nil
}

// autoMigrate 自动迁移表结构
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Event{},
		&model.Session{},
		&model.SessionDiff{},
		&model.SkillNode{},
		&model.Diff{},
		&model.DailySummary{},
		&model.PeriodSummary{},
		&model.BrowserEvent{},
	)
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
