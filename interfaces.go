package interfaces

type MigrationTool interface {
	IsMigrationNeeded() (bool, error)
	PreMigrate() error
	Migrate() error
	PostMigrate() error
}
