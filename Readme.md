# Golang in-memory key-value storage with time-to-life

## Create new map

```go
import ttl_map "github.com/leprosus/golang-ttl-map"

ttlMap := ttl_map.New("/path/to/auto-save-file.tsv")
```

## Set/Get

```go
ttlMap.Set("key", "value", 60)

value := ttlMap.Get("key")
```

## Save/Restore

```go
ttlMap.Save()

ttlMap.Restore()
```

## List all methods


* New(filePath) - creates new map
* Set(key, value, ttl) - adds value by key with ttl in seconds
* Get(key) - returns value or empty string
* Del(key) - deletes value by key
* Save() - saves map in tsv file
* Restore() - restores map from tsv file