package db

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/lightningnetwork/lnd/sqldb/v2"
)

const (
	// LatestMigrationVersion is the latest migration version of the
	// database. This is used to implement downgrade protection for the
	// daemon.
	//
	// NOTE: This MUST be updated when a new migration is added.
	LatestMigrationVersion = 5
)

// MakeTestMigrationStreams creates the migration streams for the unit test
// environment.
//
// NOTE: This function is not located in the migrationstreams package to avoid
// cyclic dependencies.
func MakeTestMigrationStreams() []sqldb.MigrationStream {
	migStream := sqldb.MigrationStream{
		MigrateTableName: pgx.DefaultMigrationsTable,
		SQLFileDirectory: "sqlc/migrations",
		Schemas:          SqlSchemas,

		// LatestMigrationVersion is the latest migration version of the
		// database.  This is used to implement downgrade protection for
		// the daemon.
		//
		// NOTE: This MUST be updated when a new migration is added.
		LatestMigrationVersion: LatestMigrationVersion,

		MakePostMigrationChecks: func(
			db *sqldb.BaseDB) (map[uint]migrate.PostStepCallback,
			error) {

			return make(map[uint]migrate.PostStepCallback), nil
		},
	}

	return []sqldb.MigrationStream{migStream}
}
