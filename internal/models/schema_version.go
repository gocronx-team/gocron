package models

// SchemaVersion stores the database migration level for managed deployments,
// where local .version files are intentionally ephemeral.
type SchemaVersion struct {
	ID      uint `gorm:"primaryKey"`
	Version int  `gorm:"not null"`
}
