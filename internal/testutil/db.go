package testutil

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/yuqie6/WorkMirror/internal/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenTestDB 打开内存 SQLite 并自动迁移所有表
func OpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&schema.Event{},
		&schema.Session{},
		&schema.SkillNode{},
		&schema.Diff{},
		&schema.DailySummary{},
		&schema.BrowserEvent{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}
