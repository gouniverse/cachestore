# Cache Store

Cache messages to a database table.

## Installation
```
go get -u github.com/gouniverse/cachestore
```

## Setup

```
cacheStore = cachestore.NewStore(cachestore.WithGormDb(databaseInstance), cachestore.WithTableName("my_cache"))

go cacheStore.ExpireCacheGoroutine()
```

## Usage

- Set value to cache with expiration
```
isSaved := cacheStore.Set("token", "ABCDEFGHIJKLMNOPQRSTVUXYZ", 60*60) // 1 hour
if isSaved == false {
		log.Println("Saving failed")
		return
}
```

- Get value from cache with default if not found
```
token := cacheStore.Get("token", "")
```

## Changelog
2021.09.11 - Removed GORM dependency and moved to the standard library
