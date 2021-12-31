# Cache Store

[![Tests Status](https://github.com/gouniverse/cachestore/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/gouniverse/cachestore/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gouniverse/cachestore)](https://goreportcard.com/report/github.com/gouniverse/cachestore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gouniverse/cachestore)](https://pkg.go.dev/github.com/gouniverse/cachestore)

Cache messages to a database table.

## Installation
```
go get -u github.com/gouniverse/cachestore
```

## Setup

```
cacheStore = cachestore.NewStore(cachestore.WithDb(databaseInstance), cachestore.WithTableName("my_cache"), cachestore.WithDebug(true))

go cacheStore.ExpireCacheGoroutine()
```

## Usage

- Set value to cache with expiration
```
isSaved, err := cacheStore.Set("token", "ABCDEFGHIJKLMNOPQRSTVUXYZ", 60*60) // 1 hour (= 60 min * 60 sec)
if isSaved == false {
	log.Println("Saving failed")
	return
}
```

- Get value from cache with default if not found
```
token, err := cacheStore.Get("token", "") // "" - default value, if the key has expired, or missing
```

- Set and retrieve complex value as JSON
```
isSaved, err := cacheStore.Set("token", map[string]string{"first_name": "Jo"}, 60*60) // 1 hour (= 60 min * 60 sec)
if isSaved == false {
	log.Println("Saving failed")
	return
}

value, err := store.GetJSON("hello", "")

if err != nil {
	log.Fatalf("Getting JSON failed:" + err.Error())
}

result := value.(map[string]interface{})

log.Println(result["first_name"])
```

## Changelog

2021.12.31 - Fixed GetJSON and added tests

2021.12.29 - Cache ID updated to nano precission

2021.12.27 - Cache key length increased

2021.12.12 - Added license

2021.12.12 - Added tests badge

2021.12.12 - Fixed bug where DB scanner was returning empty values

2021.12.09 - Added support for DB dialects

2021.09.11 - Removed GORM dependency and moved to the standard library
