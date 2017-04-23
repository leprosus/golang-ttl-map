package ttl_map

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type data struct {
	key       string
	value     string
	timestamp int64
}

type Heap struct {
	sync.RWMutex
	data     map[string]data
	filePath string
}

func New(filePath string) Heap {
	heap := Heap{
		data:     map[string]data{},
		filePath: filePath}

	return heap
}

func (heap *Heap) append(one data) {
	file, err := os.OpenFile(heap.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return
	}
	defer file.Sync()
	defer file.Close()

	writer := bufio.NewWriter(file)
	if one.timestamp < time.Now().Unix() {
		return
	}

	line := fmt.Sprintf("%s\t%s\t%d\n", one.key, one.value, one.timestamp)

	if num, err := writer.WriteString(line); err == nil && num < len(line) {
		return
	}
	writer.Flush()
}

func (heap *Heap) Set(key string, value string, ttl int64) {
	one := data{
		value:     value,
		timestamp: time.Now().Unix() + ttl}

	heap.Lock()
	heap.data[key] = one
	heap.Unlock()

	one.key = key
	go heap.append(one)
}

func (heap *Heap) Get(key string) string {
	heap.RLock()
	one, ok := heap.data[key]
	heap.RUnlock()

	if ok {
		if one.timestamp <= time.Now().Unix() {
			heap.Del(key)

			return ""
		} else {
			return one.value
		}
	} else {
		return ""
	}

	return ""
}

func (heap *Heap) Del(key string) {
	heap.Lock()
	delete(heap.data, key)
	heap.Unlock()
}

func (heap *Heap) Save() {
	os.Remove(heap.filePath)

	heap.RLock()
	for key, one := range heap.data {
		if one.timestamp > time.Now().Unix() {
			one.key = key
			go heap.append(one)
		}
	}
	heap.RUnlock()
}

func (heap *Heap) Restore() bool {
	_, err := os.Stat(heap.filePath)
	if err != nil {
		return false
	}

	file, err := os.OpenFile(heap.filePath, os.O_RDONLY, 0777)
	if err != nil {
		return false
	}
	defer file.Close()

	heap.Lock()
	defer heap.Unlock()

	heap.data = map[string]data{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		slices := strings.Split(line, "\t")
		if len(slices) == 3 {
			key := slices[0]
			value := slices[1]

			timestamp, err := strconv.ParseInt(slices[2], 10, 64)
			if err != nil {
				return false
			}

			heap.data[key] = data{
				value:     value,
				timestamp: timestamp}
		}
	}

	return true
}
