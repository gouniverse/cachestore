package cachestore

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		panic(err)
	}

	return db
}

func TestStoreCreate(t *testing.T) {
	db := InitDB("test_cache_store_create.db")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("Store could not be created: " + err.Error())
	}

	if store == nil {
		t.Fatalf("Store could not be created")
	}

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

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_automigrate",
		AutomigrateEnabled: false,
	})

	err := store.AutoMigrate()

	if err != nil {
		t.Fatalf("Automigrate failed: " + err.Error())
	}

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

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_automigrate",
		AutomigrateEnabled: true,
	})

	err := store.Remove("post")

	if err != nil {
		t.Fatalf("Entiry could not be created: " + err.Error())
	}

	val, err := store.FindByKey("post")
	if err != nil {
		t.Fatalf("Getting JSON failed:" + err.Error())
	}
	if val != nil {
		t.Fatalf("Cache should no longer be present")
	}
}

func TestStoreEnableDebug(t *testing.T) {
	db := InitDB("test_cache_debug.db")

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_debug",
		AutomigrateEnabled: false,
	})
	store.EnableDebug(true)

	err := store.AutoMigrate()

	if err != nil {
		t.Fatalf("Automigrate failed: " + err.Error())
	}
}

func TestSetKey(t *testing.T) {
	db := InitDB("test_cache_set_key.db")

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_set_key",
		AutomigrateEnabled: true,
	})

	ok, err := store.Set("hello", "world", 1)

	if err != nil {
		t.Fatalf("Setting key failed: " + err.Error())
	}

	if ok != true {
		t.Fatalf("Response not true: " + err.Error())
	}

	value, err := store.Get("hello", "")
	if err != nil {
		t.Fatalf("Getting JSON failed:" + err.Error())
	}

	if value != "world" {
		t.Fatalf("Incorrect value: " + err.Error())
	}
}

func TestUpdateKey(t *testing.T) {
	db := InitDB("test_cache_update_key.db")

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_update_key",
		AutomigrateEnabled: true,
	})

	ok, err := store.Set("hello", "world", 1)

	if err != nil {
		t.Fatalf("Setting key failed: " + err.Error())
	}

	if ok != true {
		t.Fatalf("Response not true: " + err.Error())
	}

	cache1, err := store.FindByKey("hello")

	if err != nil {
		t.Fatalf("Find by key failed:" + err.Error())
	}

	time.Sleep(2 * time.Second)

	ok2, err2 := store.Set("hello", "world", 1)

	if err2 != nil {
		t.Fatalf("Update setting key failed: " + err2.Error())
	}

	if ok2 != true {
		t.Fatalf("Update response not true: " + err.Error())
	}

	cache2, err := store.FindByKey("hello")
	if err != nil {
		t.Fatalf("Find by key failed:" + err.Error())
	}

	if cache2 == nil {
		t.Fatalf("Cache not found: " + err.Error())
	}

	if cache2.Value != "world" {
		t.Fatalf("Value not correct: " + cache2.Value)
	}

	if cache2.Key != "hello" {
		t.Fatalf("Key not correct: " + cache2.Key)
	}

	if cache2.UpdatedAt == cache1.CreatedAt {
		t.Fatalf("Updated at should be different from created at date: " + cache2.UpdatedAt.Format(time.UnixDate))
	}

	if cache2.UpdatedAt.Sub(cache1.CreatedAt).Seconds() < 1 {
		t.Fatalf("Updated at should more than 1 second after created at date: " + cache2.UpdatedAt.Format(time.UnixDate) + " - " + cache1.CreatedAt.Format(time.UnixDate))
	}
}

func TestSetGetJSON(t *testing.T) {
	db := InitDB("test_cache_set_json.db")

	store, _ := NewStore(NewStoreOptions{
		DB:                 db,
		CacheTableName:     "cache_automigrate",
		AutomigrateEnabled: true,
	})

	ok, err := store.SetJSON("hello", map[string]string{"first_name": "Jo"}, 1)

	if err != nil {
		t.Fatalf("Setting key failed: " + err.Error())
	}

	if ok != true {
		t.Fatalf("Response not true: " + err.Error())
	}

	value, err := store.GetJSON("hello", "")

	if err != nil {
		t.Fatalf("Getting JSON failed:" + err.Error())
	}

	result := value.(map[string]interface{})
	if result["first_name"] != "Jo" {
		t.Fatalf("Incorrect value: %s", value)
	}
}
