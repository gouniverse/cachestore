package cachestore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"     // importing mysql dialect
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"  // importing postgres dialect
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"   // importing sqlite3 dialect
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver" // importing sqlserver dialect
	"github.com/georgysavva/scany/sqlscan"
	"github.com/gouniverse/uid"
)

// Store defines a session store
type storeImplementation struct {
	cacheTableName     string
	dbDriverName       string
	db                 *sql.DB
	automigrateEnabled bool
	debugEnabled       bool
}

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	CacheTableName     string
	DB                 *sql.DB
	DbDriverName       string
	TimeoutSeconds     int64
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	store := &storeImplementation{
		cacheTableName:     opts.CacheTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.cacheTableName == "" {
		return nil, errors.New("cache store: sessionTableName is required")
	}

	if store.db == nil {
		return nil, errors.New("cache store: DB is required")
	}

	if store.dbDriverName == "" {
		store.dbDriverName = store.DriverName(store.db)
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store, nil
}

// AutoMigrate auto migrate
func (st *storeImplementation) AutoMigrate() error {
	sql := st.SQLCreateTable()

	_, err := st.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// DriverName finds the driver name from database
func (st *storeImplementation) DriverName(db *sql.DB) string {
	dv := reflect.ValueOf(db.Driver())

	driverFullName := dv.Type().String()

	if strings.Contains(driverFullName, "mysql") {
		return "mysql"
	}

	if strings.Contains(driverFullName, "postgres") || strings.Contains(driverFullName, "pq") {
		return "postgres"
	}

	if strings.Contains(driverFullName, "sqlite") {
		return "sqlite"
	}

	if strings.Contains(driverFullName, "mssql") {
		return "mssql"
	}

	return driverFullName
}

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debugEnabled bool) {
	st.debugEnabled = debugEnabled
}

// ExpireCacheGoroutine - soft deletes expired cache
func (st *storeImplementation) ExpireCacheGoroutine() error {
	i := 0
	for {
		i++
		if st.debugEnabled {
			log.Println("Cleaning expired cache...")
		}
		sqlStr, _, errSql := goqu.Dialect(st.dbDriverName).From(st.cacheTableName).Where(goqu.C("expires_at").Lt(time.Now())).Delete().ToSQL()

		if errSql != nil {
			if st.debugEnabled {
				log.Println(errSql.Error())
			}
			return errSql
		}

		if st.debugEnabled {
			log.Println(sqlStr)
		}

		_, err := st.db.Exec(sqlStr)
		if err != nil {
			if err == sql.ErrNoRows {
				// Looks like this is now outdated for sqlscan
				return nil
			}
			if sqlscan.NotFound(err) {
				return nil
			}
			log.Println("CacheStore. ExpireCacheGoroutine. Error: ", err)
			return nil
		}

		time.Sleep(60 * time.Second) // Every minute
	}
}

// FindByKey finds a cache by key
func (st *storeImplementation) FindByKey(key string) (*Cache, error) {
	sqlStr, _, errSql := goqu.Dialect(st.dbDriverName).
		From(st.cacheTableName).
		Where(goqu.C("cache_key").Eq(key), goqu.C("deleted_at").IsNull()).
		Select("*").
		Limit(1).
		ToSQL()

	if errSql != nil {
		if st.debugEnabled {
			log.Println(errSql.Error())
		}
		return nil, errSql
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	var cache Cache
	err := sqlscan.Get(context.Background(), st.db, &cache, sqlStr)

	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil, nil
		}
		if sqlscan.NotFound(err) {
			return nil, nil
		}

		if st.debugEnabled {
			log.Println("CacheStore. FindByKey. Error: ", err)
		}

		return nil, err
	}

	return &cache, nil
}

// Get gets a key from cache
func (st *storeImplementation) Get(key string, valueDefault string) (string, error) {
	cache, errFind := st.FindByKey(key)

	if errFind != nil {
		return valueDefault, errFind
	}

	if cache != nil {
		return cache.Value, nil
	}

	return valueDefault, nil
}

// GetJSON gets a JSON key from cache
func (st *storeImplementation) GetJSON(key string, valueDefault interface{}) (interface{}, error) {
	cache, errFind := st.FindByKey(key)

	if errFind != nil {
		return valueDefault, errFind
	}

	if cache != nil {
		jsonValue := cache.Value
		var e interface{}
		jsonError := json.Unmarshal([]byte(jsonValue), &e)
		if jsonError != nil {
			return valueDefault, jsonError
		}

		return e, nil
	}

	return valueDefault, nil
}

// Remove removes a key from cache
func (st *storeImplementation) Remove(key string) error {
	sqlStr, _, errSql := goqu.Dialect(st.dbDriverName).
		From(st.cacheTableName).
		Where(goqu.C("cache_key").Eq(key), goqu.C("deleted_at").IsNull()).
		Delete().
		ToSQL()

	if errSql != nil {
		if st.debugEnabled {
			log.Println(errSql.Error())
		}
		return errSql
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// Looks like this is now outdated for sqlscan
			return nil
		}

		if sqlscan.NotFound(err) {
			return nil
		}

		log.Println("CacheStore. Error: ", err)
		return nil
	}

	return nil
}

// Set sets new key value pair
func (st *storeImplementation) Set(key string, value string, seconds int64) error {
	cache, errFind := st.FindByKey(key)

	if errFind != nil {
		return errFind
	}

	expiresAt := time.Now().Add(time.Second * time.Duration(seconds))

	var sqlStr string
	var errSql error
	if cache == nil {
		var newCache = Cache{
			ID:        uid.NanoUid(),
			Key:       key,
			Value:     value,
			ExpiresAt: &expiresAt,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		sqlStr, _, errSql = goqu.Dialect(st.dbDriverName).
			Insert(st.cacheTableName).
			Rows(newCache).
			ToSQL()
	} else {
		fields := map[string]interface{}{}
		fields["cache_value"] = value
		fields["expires_at"] = &expiresAt
		fields["updated_at"] = time.Now()

		sqlStr, _, errSql = goqu.Dialect(st.dbDriverName).
			Update(st.cacheTableName).
			Set(fields).
			Where(goqu.C("id").Eq(cache.ID)).
			ToSQL()
	}

	if errSql != nil {
		if st.debugEnabled {
			log.Println(errSql.Error())
		}
		return errSql
	}

	if st.debugEnabled {
		log.Println(sqlStr)
	}

	_, errExec := st.db.Exec(sqlStr)

	if errExec != nil {
		return errExec
	}

	return nil
}

// SetJSON sets new key JSON value pair
func (st *storeImplementation) SetJSON(key string, value interface{}, seconds int64) error {
	jsonValue, jsonError := json.Marshal(value)

	if jsonError != nil {
		return jsonError
	}

	return st.Set(key, string(jsonValue), seconds)
}

// SQLCreateTable returns a SQL string for creating the cache table
func (st *storeImplementation) SQLCreateTable() string {
	sqlMysql := `
	CREATE TABLE IF NOT EXISTS ` + st.cacheTableName + ` (
	  id varchar(40) NOT NULL PRIMARY KEY,
	  cache_key varchar(255) NOT NULL,
	  cache_value text,
	  expires_at datetime,
	  created_at datetime NOT NULL,
	  updated_at datetime NOT NULL,
	  deleted_at datetime
	);
	`

	sqlPostgres := `
	CREATE TABLE IF NOT EXISTS "` + st.cacheTableName + `" (
	  "id" varchar(40) NOT NULL PRIMARY KEY,
	  "cache_key" varchar(255) NOT NULL,
	  "cache_value" text,
	  "expires_at" timestamptz(6),
	  "created_at" timestamptz(6) NOT NULL,
	  "updated_at" timestamptz(6) NOT NULL,
	  "deleted_at" timestamptz(6)
	)
	`

	sqlSqlite := `
	CREATE TABLE IF NOT EXISTS "` + st.cacheTableName + `" (
	  "id" varchar(40) NOT NULL PRIMARY KEY,
	  "cache_key" varchar(255) NOT NULL,
	  "cache_value" text,
	  "expires_at" datetime,
	  "created_at" datetime NOT NULL,
	  "updated_at" datetime NOT NULL,
	  "deleted_at" datetime
	)
	`

	sql := "unsupported driver " + st.dbDriverName

	if st.dbDriverName == "mysql" {
		sql = sqlMysql
	}
	if st.dbDriverName == "postgres" {
		sql = sqlPostgres
	}
	if st.dbDriverName == "sqlite" {
		sql = sqlSqlite
	}

	return sql
}
