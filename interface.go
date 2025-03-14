package cachestore

import "database/sql"

type StoreInterface interface {
	AutoMigrate() error
	EnableDebug(debugEnabled bool)
	DriverName(db *sql.DB) string

	ExpireCacheGoroutine() error

	// SQLCreateTable() string

	Set(key string, value string, seconds int64) error
	Get(key string, valueDefault string) (string, error)
	SetJSON(key string, value any, seconds int64) error
	GetJSON(key string, valueDefault any) (any, error)
	Remove(key string) error
	FindByKey(key string) (*Cache, error)
}
