package cachestore

import (
	//"log"
	// "log"
	"database/sql"
	"testing"

	//"database/sql"
	_ "github.com/mattn/go-sqlite3"
	// _ "modernc.org/sqlite"
)

func InitDB(filepath string) *sql.DB {
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		panic(err)
	}

	return db
}

func TestStoreCreate(t *testing.T) {
	db := InitDB("test_cache_store_create.db")

	store, _ := NewStore(WithDb(db), WithTableName("cache"), WithAutoMigrate(true))

	isOk, err := store.Set("post", "1234567890", 5)

	if err != nil {
		t.Fatalf("Cache could not be created: " + err.Error())
	}

	if isOk == false {
		t.Fatalf("Cache could not be created")
	}
}

func TestStoreAutomigrate(t *testing.T) {
	db := InitDB("test_cache_automigrate.db")

	store, _ := NewStore(WithDb(db), WithTableName("cache_automigrate"))

	store.AutoMigrate()

	isOk, err := store.Set("post", "1234567890", 5)

	if err != nil {
		t.Fatalf("Cache could not be created: " + err.Error())
	}

	if isOk == false {
		t.Fatalf("Cache could not be created")
	}
}

func TestStoreCacheDelete(t *testing.T) {
	db := InitDB("test_cache_delete.db")

	store, _ := NewStore(WithDb(db), WithTableName("cache"), WithAutoMigrate(true))

	err := store.Remove("post")

	if err != nil {
		t.Fatalf("Entiry could not be created: " + err.Error())
	}

	if store.FindByKey("post") != nil {
		t.Fatalf("Cache should no longer be present")
	}
}