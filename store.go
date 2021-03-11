package cachestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Store defines a session store
type Store struct {
	cacheTableName     string
	db                 *gorm.DB
	automigrateEnabled bool
}

// StoreOption options for the cache store
type StoreOption func(*Store)

// WithAutoMigrate sets the table name for the cache store
func WithAutoMigrate(automigrateEnabled bool) StoreOption {
	return func(s *Store) {
		s.automigrateEnabled = automigrateEnabled
	}
}

// WithDriverAndDNS sets the driver and the DNS for the database for the cache store
func WithDriverAndDNS(driverName string, dsn string) StoreOption {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	return func(s *Store) {
		s.db = db
	}
}

// WithGormDb sets the GORM database for the cache store
func WithGormDb(db *gorm.DB) StoreOption {
	return func(s *Store) {
		s.db = db
	}
}

// WithTableName sets the table name for the cache store
func WithTableName(cacheTableName string) StoreOption {
	return func(s *Store) {
		s.cacheTableName = cacheTableName
	}
}

// NewStore creates a new entity store
func NewStore(opts ...StoreOption) *Store {
	store := &Store{}
	for _, opt := range opts {
		opt(store)
	}

	if store.cacheTableName == "" {
		log.Panic("User store: cacheTableName is required")
	}

	if store.automigrateEnabled == true {
		store.AutoMigrate()
	}

	return store
}

// AutoMigrate auto migrate
func (st *Store) AutoMigrate() {
	st.db.Table(st.cacheTableName).AutoMigrate(&Cache{})
}

// ExpireCacheGoroutine - soft deletes expired cache
func (st *Store) ExpireCacheGoroutine() {
	i := 0
	for {
		i++
		fmt.Println("Cleaning expired sessions...")
		st.db.Table(st.cacheTableName).Where("`expires_at` < ?", time.Now()).Delete(Cache{})
		time.Sleep(60 * time.Second) // Every minute
	}
}

// FindByKey finds a cache by key
func (st *Store) FindByKey(key string) *Cache {
	// log.Println(key)

	cache := &Cache{}

	result := st.db.Table(st.cacheTableName).Where("`cache_key` = ?", key).First(&cache)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}

		log.Panic(result.Error)
	}

	return cache
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

// GetJSON gets a JSON key from cache
func (st *Store) Remove(key string) {
	st.db.Table(st.cacheTableName).Where("`cache_key` = ?", key).Delete(Cache{})
}

// Set sets new key value pair
func (st *Store) Set(key string, value string, seconds int64) bool {
	cache := st.FindByKey(key)

	expiresAt := time.Now().Add(time.Second * time.Duration(seconds))

	if cache != nil {
		cache.Value = value
		cache.ExpiresAt = &expiresAt
		//dbResult := GetDb().Table(User).Where("`key` = ?", key).Update(&cache)
		dbResult := st.db.Table(st.cacheTableName).Save(&cache)
		if dbResult != nil {
			return false
		}
		return true
	}

	var newCache = Cache{Key: key, Value: value, ExpiresAt: &expiresAt}

	dbResult := st.db.Table(st.cacheTableName).Create(&newCache)

	if dbResult.Error != nil {
		return false
	}

	return true
	// sql := sb.NewSqlite().Table(st.tableName).Insert(map[string]string{
	// 	"id":          uid.NanoUid(),
	// 	"cache_key":   key,
	// 	"cache_value": value,
	// 	"expires_at":  expiresAt.Format("2006-01-02T15:04:05"),
	// 	"created_at":  time.Now().Format("2006-01-02T15:04:05"),
	// 	"updated_at":  time.Now().Format("2006-01-02T15:04:05"),
	// })
	// log.Println(sql)
	// return true
}

// Set sets new key value pair
func (st *Store) SetJSON(key string, value interface{}, seconds int64) bool {
	jsonValue, jsonError := json.Marshal(value)
	if jsonError != nil {
		return false
	}

	cache := st.FindByKey(key)

	expiresAt := time.Now().Add(time.Second * time.Duration(seconds))

	if cache != nil {

		cache.Value = string(jsonValue)
		cache.ExpiresAt = &expiresAt
		//dbResult := GetDb().Table(User).Where("`key` = ?", key).Update(&cache)
		dbResult := st.db.Table(st.cacheTableName).Save(&cache)
		if dbResult != nil {
			return false
		}
		return true
	}

	var newCache = Cache{Key: key, Value: string(jsonValue), ExpiresAt: &expiresAt}

	dbResult := st.db.Table(st.cacheTableName).Create(&newCache)

	if dbResult.Error != nil {
		return false
	}

	return true
	// sql := sb.NewSqlite().Table(st.tableName).Insert(map[string]string{
	// 	"id":          uid.NanoUid(),
	// 	"cache_key":   key,
	// 	"cache_value": value,
	// 	"expires_at":  expiresAt.Format("2006-01-02T15:04:05"),
	// 	"created_at":  time.Now().Format("2006-01-02T15:04:05"),
	// 	"updated_at":  time.Now().Format("2006-01-02T15:04:05"),
	// })
	// log.Println(sql)
	// return true
}
