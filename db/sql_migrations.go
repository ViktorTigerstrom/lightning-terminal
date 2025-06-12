package db

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/lightningnetwork/lnd/sqldb/v2"
)

func MakeMigrationStreams(
	extraChecks map[uint]PostMigrationChecker) []sqldb.MigrationStream {

	checks := postMigrationChecks

	for version, check := range extraChecks {
		// Add or overwrite the checks map with the extra check.
		checks[version] = check
	}

	migStream := sqldb.MigrationStream{
		MigrateTableName: pgx.DefaultMigrationsTable,
		SQLFileDirectory: "sqlc/migrations",
		Schemas:          sqlSchemas,

		// LatestMigrationVersion is the latest migration version of the
		// database.  This is used to implement downgrade protection for
		// the daemon.
		//
		// NOTE: This MUST be updated when a new migration is added.
		LatestMigrationVersion: LatestMigrationVersion,

		MakePostMigrationChecks: func(
			db *sqldb.BaseDB) (map[uint]migrate.PostStepCallback,
			error) {

			return makePostStepCallbacks(db, checks),
				nil
		},
	}

	return []sqldb.MigrationStream{migStream}
}
