package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // 纯 Go SQLite 驱动
	"github.com/yuqie6/WorkMirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	DB             *gorm.DB
	SafeMode       bool
	SchemaVersion  int
	MigrationError string
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

	d := &Database{DB: db}
	if err := migrateWithVersion(db, d); err != nil {
		// v0.2 产品化：迁移失败进入“安全模式”，允许 UI 启动并导出诊断信息。
		d.SafeMode = true
		d.MigrationError = err.Error()
		slog.Error("数据库迁移失败，进入安全模式", "error", err)
	}

	slog.Info("数据库初始化成功", "path", dbPath)

	return d, nil
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
		&schema.SchemaMeta{},
		&schema.Event{},
		&schema.Session{},
		&schema.SessionDiff{},
		&schema.SkillNode{},
		&schema.SkillActivity{},
		&schema.Diff{},
		&schema.DailySummary{},
		&schema.PeriodSummary{},
		&schema.BrowserEvent{},
	)
}

const latestSchemaVersion = 1

func migrateWithVersion(db *gorm.DB, out *Database) error {
	if db == nil {
		return fmt.Errorf("db 不能为空")
	}
	if out == nil {
		return fmt.Errorf("out 不能为空")
	}

	// 先确保 schema_meta 存在（即使后续迁移失败，也能记录状态）
	if err := db.AutoMigrate(&schema.SchemaMeta{}); err != nil {
		return fmt.Errorf("创建 schema_meta 失败: %w", err)
	}

	var meta schema.SchemaMeta
	err := db.First(&meta, 1).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			meta = schema.SchemaMeta{ID: 1, SchemaVersion: 0}
			if err := db.Create(&meta).Error; err != nil {
				return fmt.Errorf("初始化 schema_meta 失败: %w", err)
			}
		} else {
			return fmt.Errorf("读取 schema_meta 失败: %w", err)
		}
	}

	cur := meta.SchemaVersion
	out.SchemaVersion = cur

	if cur > latestSchemaVersion {
		return fmt.Errorf("数据库 schema_version=%d 高于当前程序支持的版本=%d", cur, latestSchemaVersion)
	}
	if cur == latestSchemaVersion {
		return nil
	}

	// v0.2：迁移策略保持最小化——仍基于 AutoMigrate，但以 schema_version 作为升级门闸。
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("迁移数据库失败: %w", err)
	}

	meta.SchemaVersion = latestSchemaVersion
	if err := db.Save(&meta).Error; err != nil {
		return fmt.Errorf("写入 schema_meta 失败: %w", err)
	}
	out.SchemaVersion = latestSchemaVersion
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
