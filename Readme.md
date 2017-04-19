# Golang in-memory key-value storage with time-to-life

## Create new map

```go
import ttl_map "github.com/leprosus/golang-ttl-map"

ttlMap := ttl_map.New()
```

## Set/Get

```go
ttlMap.Set("key", "value", 60)

value := ttlMap.Get("key")
```

## Save/Restore

```go
ttlMap.Save("/path/to/file.tsv")

ttlMap.Restore("/path/to/file.tsv")
```

## List all methods


* New() - creates new map
* Set(key, value, ttl) - adds value by key with ttl in seconds
* Get(key) - returns value or empty string
* Del(key) - deletes value by key
* Save(filePath) - saves map in tsv file
* Restore(filePath) - restores map from tsv file