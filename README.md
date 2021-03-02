# cachestore


## Usage

```
cacheStore = cachestore.NewStore(cachestore.WithGormDb(databaseInstance), cachestore.WithTableName("my_cache"))

go cacheStore.ExpireCacheGoroutine()
```
