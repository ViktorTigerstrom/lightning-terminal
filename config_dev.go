//go:build dev

package terminal

import (
	"context"
	"fmt"
	"github.com/lightninglabs/lightning-terminal/db/migrationstreams"
	"path/filepath"

	"github.com/lightninglabs/lightning-terminal/accounts"
	"github.com/lightninglabs/lightning-terminal/db"
	"github.com/lightninglabs/lightning-terminal/db/sqlc"
	"github.com/lightninglabs/lightning-terminal/firewalldb"
	"github.com/lightninglabs/lightning-terminal/session"
	"github.com/lightningnetwork/lnd/clock"
	"github.com/lightningnetwork/lnd/sqldb/v2"
)

const (
	// DatabaseBackendSqlite is the name of the SQLite database backend.
	DatabaseBackendSqlite = "sqlite"

	// DatabaseBackendPostgres is the name of the Postgres database backend.
	DatabaseBackendPostgres = "postgres"

	// DatabaseBackendBbolt is the name of the bbolt database backend.
	DatabaseBackendBbolt = "bbolt"

	// defaultSqliteDatabaseFileName is the default name of the SQLite
	// database file.
	defaultSqliteDatabaseFileName = "litd.db"
)

// defaultSqliteDatabasePath is the default path under which we store
// the SQLite database file.
var defaultSqliteDatabasePath = filepath.Join(
	DefaultLitDir, DefaultNetwork, defaultSqliteDatabaseFileName,
)

// DevConfig is a struct that holds the configuration options for a development
// environment. The purpose of this struct is to hold config options for
// features not yet available in production. Since our itests are built with
// the dev tag, we can test these features in our itests.
//
// nolint:lll
type DevConfig struct {
	// DatabaseBackend is the database backend we will use for storing all
	// account related data. While this feature is still in development, we
	// include the bbolt type here so that our itests can continue to be
	// tested against a bbolt backend. Once the full bbolt to SQL migration
	// is complete, however, we will remove the bbolt option.
	DatabaseBackend string `long:"databasebackend" description:"The database backend to use for storing all account related data." choice:"bbolt" choice:"sqlite" choice:"postgres"`

	// Sqlite holds the configuration options for a SQLite database
	// backend.
	Sqlite *db.SqliteConfig `group:"sqlite" namespace:"sqlite"`

	// Postgres holds the configuration options for a Postgres database
	Postgres *db.PostgresConfig `group:"postgres" namespace:"postgres"`

	MigrateToSql bool `long:"migrate-to-sql" description:"Migrate the bbolt database to SQL."`
}

// Validate checks that all the values set in our DevConfig are valid and uses
// the passed parameters to override any defaults if necessary.
func (c *DevConfig) Validate(dbDir, network string) error {
	// We'll update the database file location if it wasn't set.
	if c.Sqlite.DatabaseFileName == defaultSqliteDatabasePath {
		c.Sqlite.DatabaseFileName = filepath.Join(
			dbDir, network, defaultSqliteDatabaseFileName,
		)
	}

	if c.MigrateToSql && c.DatabaseBackend == DatabaseBackendBbolt {
		return fmt.Errorf("migrate-to-sql can only be used with an " +
			"sql backend")
	}

	return nil
}

// defaultDevConfig returns a new DevConfig with default values set.
func defaultDevConfig() *DevConfig {
	return &DevConfig{
		Sqlite: &db.SqliteConfig{
			DatabaseFileName: defaultSqliteDatabasePath,
		},
		Postgres: &db.PostgresConfig{
			Host:               "localhost",
			Port:               5432,
			MaxOpenConnections: 10,
		},
		MigrateToSql: false,
	}
}

// NewStores creates a new stores instance based on the chosen database backend.
func NewStores(ctx context.Context, cfg *Config,
	clock clock.Clock) (*stores, error) {

	var (
		networkDir = filepath.Join(cfg.LitDir, cfg.Network)
		stores     = &stores{
			closeFns: make(map[string]func() error),
		}
	)

	switch cfg.DatabaseBackend {
	case DatabaseBackendSqlite:
		// Before we initialize the SQLite store, we'll make sure that
		// the directory where we will store the database file exists.
		err := makeDirectories(networkDir)
		if err != nil {
			return stores, err
		}

		sqlStore, err := sqldb.NewSqliteStore(&sqldb.SqliteConfig{
			SkipMigrations:        cfg.Sqlite.SkipMigrations,
			SkipMigrationDbBackup: cfg.Sqlite.SkipMigrationDbBackup,
		}, cfg.Sqlite.DatabaseFileName)
		if err != nil {
			return stores, err
		}

		if !cfg.Sqlite.SkipMigrations {
			err = sqldb.ApplyAllMigrations(
				sqlStore,
				migrationstreams.MakeMigrationStreams(
					ctx, cfg.MacaroonPath, clock,
				),
			)
			if err != nil {
				return stores, fmt.Errorf("error applying "+
					"migrations to SQLlite store: %w", err,
				)
			}
		}

		queries := sqlc.NewForType(sqlStore, sqlStore.BackendType)

		acctStore := accounts.NewSQLStore(
			sqlStore.BaseDB, queries, clock,
		)
		sessStore := session.NewSQLStore(
			sqlStore.BaseDB, queries, clock,
		)
		firewallStore := firewalldb.NewSQLDB(
			sqlStore.BaseDB, queries, clock,
		)

		stores.accounts = acctStore
		stores.sessions = sessStore
		stores.firewall = firewalldb.NewDB(firewallStore)
		stores.closeFns["sqlite"] = sqlStore.BaseDB.Close

	case DatabaseBackendPostgres:
		sqlStore, err := sqldb.NewPostgresStore(&sqldb.PostgresConfig{
			Dsn:                cfg.Postgres.DSN(false),
			MaxOpenConnections: cfg.Postgres.MaxOpenConnections,
			MaxIdleConnections: cfg.Postgres.MaxIdleConnections,
			ConnMaxLifetime:    cfg.Postgres.ConnMaxLifetime,
			ConnMaxIdleTime:    cfg.Postgres.ConnMaxIdleTime,
			RequireSSL:         cfg.Postgres.RequireSSL,
			SkipMigrations:     cfg.Postgres.SkipMigrations,
		})
		if err != nil {
			return stores, err
		}

		if !cfg.Postgres.SkipMigrations {
			err = sqldb.ApplyAllMigrations(
				sqlStore, migrationstreams.MakeMigrationStreams(
					ctx, cfg.MacaroonPath, clock,
				),
			)
			if err != nil {
				return stores, fmt.Errorf("error applying "+
					"migrations to Postgres store: %w", err,
				)
			}
		}

		queries := sqlc.NewForType(sqlStore, sqlStore.BackendType)

		acctStore := accounts.NewSQLStore(
			sqlStore.BaseDB, queries, clock,
		)
		sessStore := session.NewSQLStore(
			sqlStore.BaseDB, queries, clock,
		)
		firewallStore := firewalldb.NewSQLDB(
			sqlStore.BaseDB, queries, clock,
		)

		stores.accounts = acctStore
		stores.sessions = sessStore
		stores.firewall = firewalldb.NewDB(firewallStore)
		stores.closeFns["postgres"] = sqlStore.BaseDB.Close

	default:
		accountStore, err := accounts.NewBoltStore(
			filepath.Dir(cfg.MacaroonPath), accounts.DBFilename,
			clock,
		)
		if err != nil {
			return stores, err
		}

		stores.accounts = accountStore
		stores.closeFns["bbolt-accounts"] = accountStore.Close

		sessionStore, err := session.NewDB(
			networkDir, session.DBFilename, clock, accountStore,
		)
		if err != nil {
			return stores, err
		}

		stores.sessions = sessionStore
		stores.closeFns["bbolt-sessions"] = sessionStore.Close

		firewallBoltDB, err := firewalldb.NewBoltDB(
			networkDir, firewalldb.DBFilename, stores.sessions,
			stores.accounts, clock,
		)
		if err != nil {
			return stores, fmt.Errorf("error creating firewall "+
				"BoltDB: %v", err)
		}

		stores.firewall = firewalldb.NewDB(firewallBoltDB)
		stores.closeFns["bbolt-firewalldb"] = firewallBoltDB.Close
	}

	return stores, nil
}

/*
func kvdbToSqlMigrationCallback(cfg *Config, sqlDB *sqldb.BaseDB,
	clock clock.Clock) db.PostMigrationChecker {

	var (
		writeTxOpts db.QueriesTxOptions
	)

	check := func(ctx context.Context, q6 fn.Option[*sqlcmig6.Queries],
		_ fn.Option[*sqlc.Queries]) error {

		q, err := q6.UnwrapOrErr(errors.New("sqlcmig6 queries missing"))
		if err != nil {
			return fmt.Errorf("error getting sqlcmig6 queries: %w",
				err)
		}

		tx, err := sqlDB.BeginTx(ctx, &writeTxOpts)
		if err != nil {
			return fmt.Errorf("error starting migration tx: %w",
				err)
		}

		// Ensure we roll back the migration on any error path
		defer func() {
			if err != nil {
				rollBackErr := tx.Rollback()
				if rollBackErr != nil {
					log.Errorf("error rolling back "+
						"migration tx: %v", err)
				}
			}
		}()

		accountStore, err := accounts.NewBoltStore(
			filepath.Dir(cfg.MacaroonPath), accounts.DBFilename,
			clock,
		)
		if err != nil {
			return err
		}

		err = accounts.MigrateAccountStoreToSQL(ctx, accountStore.DB, q)
		if err != nil {
			return fmt.Errorf("error migrating account store to "+
				"SQL: %v", err)
		}

		sessionStore, err := session.NewDB(
			filepath.Dir(cfg.MacaroonPath), session.DBFilename,
			clock, accountStore,
		)
		if err != nil {
			return err
		}

		err = session.MigrateSessionStoreToSQL(ctx, sessionStore.DB, q)
		if err != nil {
			return fmt.Errorf("error migrating session store to "+
				"SQL: %v", err)
		}

		firewallStore, err := firewalldb.NewBoltDB(
			filepath.Dir(cfg.MacaroonPath), firewalldb.DBFilename,
			sessionStore, accountStore, clock,
		)
		if err != nil {
			return err
		}

		err = firewalldb.MigrateFirewallDBToSQL(ctx, firewallStore.DB, q)
		if err != nil {
			return fmt.Errorf("error migrating firewalldb store "+
				"to SQL: %v", err)
		}

		// The migrations succeeded! We're therefore good to commit the tx.
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("committing migration tx: %w", err)
		}

		return nil
	}

	return *db.NewPostMigrationChecker(db.Migration6MigrateToSQL, check)
}

func KvdbToSqlMigrationCallback2(ctx context.Context, macPath string,
	sqlDB *sqldb.BaseDB, clock clock.Clock,
	q *sqlcmig6.Queries) error {

	var (
		writeTxOpts db.QueriesTxOptions
	)

	tx, err := sqlDB.BeginTx(ctx, &writeTxOpts)
	if err != nil {
		return fmt.Errorf("error starting migration tx: %w",
			err)
	}

	// Ensure we roll back the migration on any error path
	defer func() {
		if err != nil {
			rollBackErr := tx.Rollback()
			if rollBackErr != nil {
				log.Errorf("error rolling back "+
					"migration tx: %v", err)
			}
		}
	}()

	accountStore, err := accounts.NewBoltStore(
		filepath.Dir(macPath), accounts.DBFilename, clock,
	)
	if err != nil {
		return err
	}

	err = accounts.MigrateAccountStoreToSQL(ctx, accountStore.DB, q)
	if err != nil {
		return fmt.Errorf("error migrating account store to "+
			"SQL: %v", err)
	}

	sessionStore, err := session.NewDB(
		filepath.Dir(macPath), session.DBFilename,
		clock, accountStore,
	)
	if err != nil {
		return err
	}

	err = session.MigrateSessionStoreToSQL(ctx, sessionStore.DB, q)
	if err != nil {
		return fmt.Errorf("error migrating session store to "+
			"SQL: %v", err)
	}

	firewallStore, err := firewalldb.NewBoltDB(
		filepath.Dir(macPath), firewalldb.DBFilename,
		sessionStore, accountStore, clock,
	)
	if err != nil {
		return err
	}

	err = firewalldb.MigrateFirewallDBToSQL(ctx, firewallStore.DB, q)
	if err != nil {
		return fmt.Errorf("error migrating firewalldb store "+
			"to SQL: %v", err)
	}

	// The migrations succeeded! We're therefore good to commit the tx.
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing migration tx: %w", err)
	}

	return nil
}
*/
