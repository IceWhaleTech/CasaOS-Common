package interfaces

// Any logic to migrate data from previous version to current can implement this interface.
//
// The model for migrating from v0.n to v0.m is an execution chain of each version of this migration tool:
//
// START -> migration-tool-v0.n -> migration-tool-v0.n+1 -> ... -> migration-tool-v0.m -> END
//
// Therefore, each migration tool in the chain SHOULD ONLY work on data such as config files and databases.
// It is responsibility of any install/setup script to control the services because it knows when migration starts and ends.
//
// !!!IMPORTANT!!! DO NOT stop, start, enable or disable services, because the whole execution chain might not have been completed.
type MigrationTool interface {
	IsMigrationNeeded() (bool, error)
	PreMigrate() error
	Migrate() error
	PostMigrate() error
}
