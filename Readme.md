# Golang in-memory key-value storage with time-to-life

## Create new map

```go
import ttl_map "github.com/leprosus/golang-ttl-map"

heap := ttl_map.New("/path/to/auto-save-file.tsv")
```

## Set/Get

```go
heap.Set("key", "value", 60)

value := heap.Get("key")
```

## Save/Restore

```go
heap.Save()

heap.Restore()
```

## List all methods


* New(filePath) - creates new map
* Set(key, value, ttl) - adds value by key with ttl in seconds
* Get(key) - returns value or empty string
* Del(key) - deletes value by key
* Range(fn func(key string, value string, ttl int64)) - iterates all actual data
* Save() - saves map in tsv file
* Restore() - restores map from tsv file