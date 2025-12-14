package schema

import "time"

// SchemaMeta 用于记录数据库 schema 版本，避免仅依赖 AutoMigrate 导致升级不可控。
// 表内仅维护单行（ID=1）。
type SchemaMeta struct {
	ID            int       `gorm:"primaryKey"`
	SchemaVersion int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (SchemaMeta) TableName() string {
	return "schema_meta"
}
