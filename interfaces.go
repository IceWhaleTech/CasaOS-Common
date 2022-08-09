package interfaces

type Updater interface {
	IsMigrationNeeded() bool
	PreMigrate() error
	Migrate() error
	PostMigrate() error
}
