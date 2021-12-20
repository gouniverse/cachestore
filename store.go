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
type Store struct {
	cacheTableName     string
	dbDriverName       string
	db                 *sql.DB
	automigrateEnabled bool
	debug              bool
}

// StoreOption options for the cache store
type StoreOption func(*Store)

// WithAutoMigrate sets the table name for the cache store
func WithAutoMigrate(automigrateEnabled bool) StoreOption {
	return func(s *Store) {
		s.automigrateEnabled = automigrateEnabled
	}
}

// WithDb sets the database for the setting store
func WithDb(db *sql.DB) StoreOption {
	return func(s *Store) {
		s.db = db
		s.dbDriverName = s.DriverName(s.db)
	}
}

// WithDebug prints the SQL queries
func WithDebug(debug bool) StoreOption {
	return func(s *Store) {
		s.debug = debug
	}
}

// WithTableName sets the table name for the cache store
func WithTableName(cacheTableName string) StoreOption {
	return func(s *Store) {
		s.cacheTableName = cacheTableName
	}
}

// NewStore creates a new entity store
func NewStore(opts ...StoreOption) (*Store, error) {
	store := &Store{}
	for _, opt := range opts {
		opt(store)
	}

	if store.cacheTableName == "" {
		return nil, errors.New("cache store: cacheTableName is required")
	}

	if store.debug {
		log.Println(store.dbDriverName)
	}

	if store.automigrateEnabled {
		store.AutoMigrate()
	}

	return store, nil
}

// AutoMigrate auto migrate
func (st *Store) AutoMigrate() error {
	sql := st.SQLCreateTable()

	_, err := st.db.Exec(sql)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// DriverName finds the driver name from database
func (st *Store) DriverName(db *sql.DB) string {
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
func (st *Store) EnableDebug(debug bool) {
	st.debug = debug
}

// ExpireCacheGoroutine - soft deletes expired cache
func (st *Store) ExpireCacheGoroutine() error {
	i := 0
	for {
		i++
		if st.debug {
			log.Println("Cleaning expired cache...")
		}
		sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.cacheTableName).Where(goqu.C("expires_at").Lt(time.Now())).Delete().ToSQL()

		if st.debug {
			log.Println(sqlStr)
		}

		_, err := st.db.Exec(sqlStr)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			log.Fatal("Failed to execute query: ", err)
			return nil
		}

		time.Sleep(60 * time.Second) // Every minute
	}
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(key string) *Cache {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.cacheTableName).Where(goqu.C("cache_key").Eq(key), goqu.C("deleted_at").IsNull()).Select("*").ToSQL()

	if st.debug {
		log.Println(sqlStr)
	}

	var cache Cache
	err := sqlscan.Get(context.Background(), st.db, &cache, sqlStr)

	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			return nil
		}
		log.Fatal("Failed to execute query: ", err)
		return nil
	}

	return &cache
}

// Get gets a key from cache
func (st *Store) Get(key string, valueDefault string) string {
	cache := st.FindByKey(key)

	if cache != nil {
		return cache.Value
	}

	return valueDefault
}

// GetJSON gets a JSON key from cache
func (st *Store) GetJSON(key string, valueDefault interface{}) interface{} {
	cache := st.FindByKey(key)

	if cache != nil {
		jsonValue := cache.Value
		var e interface{}
		jsonError := json.Unmarshal([]byte(jsonValue), e)
		if jsonError != nil {
			return valueDefault
		}

		return e
	}

	return valueDefault
}

// Remove removes a key from cache
func (st *Store) Remove(key string) error {
	sqlStr, _, _ := goqu.Dialect(st.dbDriverName).From(st.cacheTableName).Where(goqu.C("cache_key").Eq(key), goqu.C("deleted_at").IsNull()).Delete().ToSQL()

	if st.debug {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}

		log.Fatal("Failed to execute query: ", err)
		return nil
	}

	return nil
}

// Set sets new key value pair
func (st *Store) Set(key string, value string, seconds int64) (bool, error) {
	cache := st.FindByKey(key)

	expiresAt := time.Now().Add(time.Second * time.Duration(seconds))

	var sqlStr string
	if cache == nil {
		var newCache = Cache{
			ID:        uid.MicroUid(),
			Key:       key,
			Value:     value,
			ExpiresAt: &expiresAt,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		sqlStr, _, _ = goqu.Dialect(st.dbDriverName).Insert(st.cacheTableName).Rows(newCache).ToSQL()
	} else {
		cache.Value = value
		cache.ExpiresAt = &expiresAt
		cache.UpdatedAt = time.Now()
		sqlStr, _, _ = goqu.Dialect(st.dbDriverName).Update(st.cacheTableName).Set(cache).ToSQL()
	}

	if st.debug {
		log.Println(sqlStr)
	}

	_, err := st.db.Exec(sqlStr)

	if err != nil {
		return false, err
	}

	return true, nil
}

// SetJSON sets new key value pair
func (st *Store) SetJSON(key string, value interface{}, seconds int64) (bool, error) {
	jsonValue, jsonError := json.Marshal(value)
	if jsonError != nil {
		return false, jsonError
	}

	return st.Set(key, string(jsonValue), seconds)
}

// SQLCreateTable returns a SQL string for creating the cache table
func (st *Store) SQLCreateTable() string {
	sqlMysql := `
	CREATE TABLE IF NOT EXISTS ` + st.cacheTableName + ` (
	  id varchar(40) NOT NULL PRIMARY KEY,
	  cache_key varchar(40) NOT NULL,
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
	  "cache_key" varchar(40) NOT NULL,
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
	  "cache_key" varchar(40) NOT NULL,
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
