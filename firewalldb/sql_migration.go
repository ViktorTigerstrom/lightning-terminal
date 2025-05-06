package firewalldb

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lightninglabs/lightning-terminal/db/sqlc"
	"go.etcd.io/bbolt"
)

var (
	// ErrMigrationMismatch is returned when the migrated session does not
	// match the original session.
	ErrMigrationMismatch = fmt.Errorf("migrated session does not match " +
		"original session")
)

// MigrateRulesDBToSQL runs the migration of the kv and privacy mapper stores
// from the bbolt database to a SQL database. The migration is done in a single
// transaction to ensure that all rows in the stores are migrated or none at
// all.
//
// NOTE: As sessions may contain linked accounts, the accounts sql migration
// MUST be run prior to this migration.
func MigrateRulesDBToSQL(ctx context.Context, kvStore *BoltDB, sqlDb SQLDB,
	tx SQLQueries) error {

	log.Infof("Starting migration of the rules DB to SQL")

	err := migrateKVStoresDBToSQL(ctx, kvStore, tx)
	if err != nil {
		return err
	}

	err = migratePrivacyMapperDBToSQL(ctx, kvStore, tx)
	if err != nil {
		return err
	}

	log.Infof("The rules DB has been migrated from KV to SQL.")

	return nil
}

func migrateKVStoresDBToSQL(ctx context.Context, kvStore *BoltDB,
	sqlTx SQLQueries) error {

	log.Infof("Starting migration of the KV stores to SQL")

	var totalRows int
	err := kvStore.View(func(kvTx *bbolt.Tx) error {
		for _, perm := range []bool{true, false} {
			mainBucket, err := getMainBucket(kvTx, true, perm)
			if err != nil {
				return err
			}

			err = mainBucket.ForEach(func(k, v []byte) error {
				if v != nil {
					return errors.New("expected only " +
						"buckets under main bucket")
				}

				ruleName := k
				ruleNameBucket := mainBucket.Bucket(k)
				if ruleNameBucket == nil {
					return fmt.Errorf("rule bucket %s "+
						"not found", string(k))
				}

				ruleId, err := sqlTx.GetOrInsertRuleID(
					ctx, string(ruleName),
				)
				if err != nil {
					return err
				}

				return migrateRuleBucketToSQL(
					ctx, kvTx, sqlTx, perm, ruleId,
					ruleNameBucket,
				)
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Infof("Migration of the KV stores to SQL completed. Total number "+
		"of rows migrated: %d", totalRows)

	return nil
}

func migrateRuleBucketToSQL(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64,
	ruleBucket *bbolt.Bucket) error {

	err := ruleBucket.ForEach(func(k, v []byte) error {
		if v != nil {
			return errors.New("expected only buckets under " +
				"rule-name bucket")
		}

		if bytes.Equal(k, globalKVStoreBucketKey) {
			globalBucket := ruleBucket.Bucket(
				globalKVStoreBucketKey,
			)
			if globalBucket == nil {
				return fmt.Errorf("global bucket %s for rule "+
					"id %d not found", string(k), ruleSqlId)
			}

			return migrateGlobalRuleBucketToSQL(
				ctx, kvTx, sqlTx, perm, ruleSqlId, globalBucket,
			)
		} else if bytes.Equal(k, sessKVStoreBucketKey) {
			sessionBucket := ruleBucket.Bucket(
				sessKVStoreBucketKey,
			)
			if sessionBucket == nil {
				return fmt.Errorf("session bucket %s for rule "+
					"id %d not found", string(k), ruleSqlId)
			}

			return migrateSessionBucketToSQL(
				ctx, kvTx, sqlTx, perm, ruleSqlId,
				sessionBucket,
			)
		} else {
			return fmt.Errorf("unexpected bucket %s under "+
				"rule-name bucket", string(k))
		}
	})

	return err
}

func migrateGlobalRuleBucketToSQL(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64,
	globalBucket *bbolt.Bucket) error {

	return globalBucket.ForEach(func(k, v []byte) error {
		if v == nil {
			return errors.New("expected only key-values under " +
				"global rule-name bucket")
		}

		globalInsertParams := makeGlobalInsertKVRecordParams(
			perm, ruleSqlId, string(k), v,
		)

		return sqlTx.InsertKVStoreRecord(ctx, globalInsertParams)
		// Todo: Validate inserted row
	})
}

func migrateSessionBucketToSQL(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64,
	mainSessionBucket *bbolt.Bucket) error {

	return mainSessionBucket.ForEach(func(k, v []byte) error {
		if v != nil {
			return fmt.Errorf("expected only buckets under "+
				"%s bucket", string(sessKVStoreBucketKey))
		}

		groupId := k

		groupBucket := mainSessionBucket.Bucket(groupId)
		if groupBucket == nil {
			return fmt.Errorf("group bucket for group id %s"+
				"not found", string(groupId))
		}

		return migrateSessionGroupBucketToSQL(
			ctx, kvTx, sqlTx, perm, ruleSqlId, groupId, groupBucket,
		)
	})
}

func migrateSessionGroupBucketToSQL(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64, groupAlias []byte,
	groupBucket *bbolt.Bucket) error {

	groupSqlId, err := sqlTx.GetSessionIDByAlias(
		ctx, groupAlias,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("session with group id %x "+
			"not found in sql db", groupAlias)
	} else if err != nil {
		return err
	}

	return groupBucket.ForEach(func(k, v []byte) error {
		if v == nil {
			// This is a non-feature specific k:v store for the
			// session, i.e. the session-wide store.
			sessWideParams := makeSessionWideInsertKVRecordParams(
				perm, ruleSqlId, groupSqlId, string(k), v,
			)

			return sqlTx.InsertKVStoreRecord(ctx, sessWideParams)
			// Todo: Validate inserted row
		} else if bytes.Equal(k, featureKVStoreBucketKey) {
			// This is a feature specific k:v store for the
			// session, which will be stored under the feature-name
			// under this bucket.

			featureStoreBucket := groupBucket.Bucket(
				featureKVStoreBucketKey,
			)
			if featureStoreBucket == nil {
				return fmt.Errorf("feature store bucket %s "+
					"for group id %s not found",
					string(featureKVStoreBucketKey),
					string(groupAlias))
			}

			return migrateFeatureStoreBucketToSql(
				ctx, kvTx, sqlTx, perm, ruleSqlId, groupSqlId,
				featureStoreBucket,
			)
		} else {
			return fmt.Errorf("unexpected bucket %s found under "+
				"the %s bucket", string(k),
				string(sessKVStoreBucketKey))
		}
	})
}

func migrateFeatureStoreBucketToSql(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64, groupSqlId int64,
	featureStoreBucket *bbolt.Bucket) error {

	return featureStoreBucket.ForEach(func(k, v []byte) error {
		if v != nil {
			return fmt.Errorf("expected only buckets under " +
				"feature stores bucket")
		}

		featureName := k
		featureNameBucket := featureStoreBucket.Bucket(featureName)
		if featureNameBucket == nil {
			return fmt.Errorf("feature bucket %s not found",
				string(featureName))
		}

		featureSqlId, err := sqlTx.GetOrInsertFeatureID(
			ctx, string(featureName),
		)
		if err != nil {
			return err
		}

		return migrateFeatureNameBucketToSql(
			ctx, kvTx, sqlTx, perm, ruleSqlId, groupSqlId,
			featureSqlId, featureNameBucket,
		)
	})
}

func migrateFeatureNameBucketToSql(ctx context.Context, kvTx *bbolt.Tx,
	sqlTx SQLQueries, perm bool, ruleSqlId int64, groupSqlId int64,
	featureSqlId int64, featureNameBucket *bbolt.Bucket) error {

	return featureNameBucket.ForEach(func(k, v []byte) error {
		if v == nil {
			return fmt.Errorf("expected only key-values under "+
				"feature name bucket, but found bucket %s",
				string(k))
		}

		featureParams := makeFeatureSpecificInsertKVRecordParams(
			perm, ruleSqlId, groupSqlId, featureSqlId,
			string(k), v,
		)

		return sqlTx.InsertKVStoreRecord(ctx, featureParams)
		// Todo: Validate inserted row
	})
}

func migratePrivacyMapperDBToSQL(ctx context.Context, kvStore *BoltDB,
	tx SQLQueries) error {

	log.Infof("Starting migration of the privacy mapper store to SQL")

	log.Errorf("Unimplemented migration of the privacy mapper store to SQL.")

	return nil
	/*var totalRows int

	// TODO: SET!!!! Should iterate over all group IDs
	var groupID session.ID

	testDB := kvStore.PrivacyDB(groupID)
	err := testDB.View(ctx, func(ctx context.Context, kvdbTx PrivacyMapTx) error {
		pairs, err := kvdbTx.FetchAllPairs(ctx)
		if err != nil {
			return err
		}

		pairs.mu.Lock()
		defer pairs.mu.Unlock()

		for realVal, pseudoVal := range pairs.pairs {
			tx.InsertPrivacyPair(ctx, sqlc.InsertPrivacyPairParams{
				GroupID:   int64(groupID),
				RealVal:   realVal,
				PseudoVal: pseudoVal,
			})
		}
	})
	if err != nil {
		return err
	}

	log.Infof("Migration of the privacy mapper store to SQL completed. "+
		"Total number of rows migrated: %d", totalRows)

	return nil
	*/
}

func makeGlobalInsertKVRecordParams(perm bool, ruleID int64, key string,
	value []byte) sqlc.InsertKVStoreRecordParams {

	return sqlc.InsertKVStoreRecordParams{
		EntryKey: key,
		Value:    value,
		Perm:     perm,
		RuleID:   ruleID,
	}
}

func makeSessionWideInsertKVRecordParams(perm bool, ruleID int64, groupId int64,
	key string, value []byte) sqlc.InsertKVStoreRecordParams {

	return sqlc.InsertKVStoreRecordParams{
		EntryKey: key,
		Value:    value,
		Perm:     perm,
		RuleID:   ruleID,
		SessionID: sql.NullInt64{
			Int64: groupId,
			Valid: true,
		},
	}
}

func makeFeatureSpecificInsertKVRecordParams(perm bool, ruleID int64,
	groupId int64, featureId int64, key string,
	value []byte) sqlc.InsertKVStoreRecordParams {

	return sqlc.InsertKVStoreRecordParams{
		EntryKey: key,
		Value:    value,
		Perm:     perm,
		RuleID:   ruleID,
		SessionID: sql.NullInt64{
			Int64: groupId,
			Valid: true,
		},
		FeatureID: sql.NullInt64{
			Int64: featureId,
			Valid: true,
		},
	}
}
