package interfaces

type Updater interface {
	IsMigrationNeeded() (bool, error)
	PreMigrate() error
	Migrate() error
	PostMigrate() error
}
