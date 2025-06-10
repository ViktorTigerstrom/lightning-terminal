//go:build !dev

package db

/*
import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lightningnetwork/lnd/clock"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/lightninglabs/lightning-terminal/db/sqlc"
	"github.com/lightninglabs/lightning-terminal/db/sqlcmig6"
	"github.com/lightningnetwork/lnd/fn/v2"
	"github.com/lightningnetwork/lnd/sqldb/v2"
)

const (
	// Migration6MigrateToSQL is the version of the migration that migrates
	// the Bbolt db to SQL.
	Migration6MigrateToSQL = 6
)

type PostMigrationChecker struct {
	// queriesVersion is the version of the queries that this post-migration
	// check is associated with. This is used to ensure that the correct
	// queries are used when running the post-migration check.
	queriesVersion uint
	check          postMigrationCheck
}

func NewPostMigrationChecker(queriesVersion uint,
	check postMigrationCheck) *PostMigrationChecker {

	return &PostMigrationChecker{
		queriesVersion: queriesVersion,
		check:          check,
	}
}

// postMigrationCheck is a function type for a function that performs a
// post-migration check on the database.
type postMigrationCheck func(context.Context, fn.Option[*sqlcmig6.Queries],
	fn.Option[*sqlc.Queries]) error

var (
	// postMigrationChecks is a map of functions that are run after the
	// database migration with the version specified in the key has been
	// applied. These functions are used to perform additional checks on the
	// database state that are not fully expressible in SQL.
	postMigrationChecks = map[uint]PostMigrationChecker{}
)

// makePostStepCallbacks turns the post migration checks into a map of post
// step callbacks that can be used with the migrate package. The keys of the map
// are the migration versions, and the values are the callbacks that will be
// executed after the migration with the corresponding version is applied.
func makePostStepCallbacks(db *sqldb.BaseDB,
	c map[uint]PostMigrationChecker) map[uint]migrate.PostStepCallback {

	queries := sqlc.NewForType(db, db.BackendType)
	executor := sqldb.NewTransactionExecutor(
		db, func(tx *sql.Tx) *sqlc.Queries {
			return queries.WithTx(tx)
		},
	)

	mig6queries := sqlcmig6.NewForType(db, db.BackendType)
	mig6executor := sqldb.NewTransactionExecutor(
		db, func(tx *sql.Tx) *sqlcmig6.Queries {
			return mig6queries.WithTx(tx)
		},
	)

	var (
		ctx               = context.Background()
		postStepCallbacks = make(map[uint]migrate.PostStepCallback)
	)
	for version, checker := range c {
		runCheck := func(m *migrate.Migration,
			q6 fn.Option[*sqlcmig6.Queries],
			q fn.Option[*sqlc.Queries]) error {

			log.Infof("Running post-migration check for version %d",
				version)
			start := time.Now()

			err := checker.check(ctx, q6, q)
			if err != nil {
				return fmt.Errorf("post-migration "+
					"check failed for version %d: "+
					"%w", version, err)
			}

			log.Infof("Post-migration check for version %d "+
				"completed in %v", version, time.Since(start))

			return nil
		}

		switch checker.queriesVersion {
		case Migration6MigrateToSQL:
			postStepCallbacks[version] = func(m *migrate.Migration,
				_ database.Driver) error {

				// We ignore the actual driver that's being
				// returned here, since we use
				// migrate.NewWithInstance() to create the
				// migration instance from our already
				// instantiated database backend that is also
				// passed into this function.
				return mig6executor.ExecTx(
					ctx, sqldb.NewWriteTx(),
					func(q6 *sqlcmig6.Queries) error {
						return runCheck(
							m, fn.Some(q6),
							fn.None[*sqlc.Queries](),
						)
					}, sqldb.NoOpReset,
				)
			}
		default:
			postStepCallbacks[version] = func(m *migrate.Migration,
				_ database.Driver) error {

				// We ignore the actual driver that's being
				// returned here, since we use
				// migrate.NewWithInstance() to create the
				// migration instance from our already
				// instantiated database backend that is also
				// passed into this function.
				return executor.ExecTx(
					ctx, sqldb.NewWriteTx(),
					func(q *sqlc.Queries) error {
						return runCheck(
							m,
							fn.None[*sqlcmig6.Queries](),
							fn.Some(q),
						)
					}, sqldb.NoOpReset,
				)
			}
		}
	}

	return postStepCallbacks
}

// makePostStepCallbacks turns the post migration checks into a map of post
// step callbacks that can be used with the migrate package. The keys of the map
// are the migration versions, and the values are the callbacks that will be
// executed after the migration with the corresponding version is applied.
func makePostStepCallbacksMig6(ctx context.Context, db *sqldb.BaseDB,
	macPath string, clock clock.Clock,
	migVersion uint) migrate.PostStepCallback {

	mig6queries := sqlcmig6.NewForType(db, db.BackendType)
	mig6executor := sqldb.NewTransactionExecutor(
		db, func(tx *sql.Tx) *sqlcmig6.Queries {
			return mig6queries.WithTx(tx)
		},
	)
	runCheck := func(m *migrate.Migration,
		q6 *sqlcmig6.Queries) error {

		log.Infof("Running post-migration check for version %d",
			migVersion)
		start := time.Now()

		err := kvdbToSqlMigrationCallback(
			ctx, macPath, db, clock, q6,
		)
		if err != nil {
			return fmt.Errorf("post-migration "+
				"check failed for version %d: "+
				"%w", migVersion, err)
		}

		log.Infof("Post-migration check for version %d "+
			"completed in %v", migVersion, time.Since(start))

		return nil
	}

	return func(m *migrate.Migration,
		_ database.Driver) error {

		// We ignore the actual driver that's being
		// returned here, since we use
		// migrate.NewWithInstance() to create the
		// migration instance from our already
		// instantiated database backend that is also
		// passed into this function.
		return mig6executor.ExecTx(
			ctx, sqldb.NewWriteTx(),
			func(q6 *sqlcmig6.Queries) error {
				return runCheck(m, q6)
			}, sqldb.NoOpReset,
		)
	}
}
*/
