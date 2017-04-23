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
	queue    chan data
}

func New(filePath string) Heap {
	heap := Heap{
		data:     map[string]data{},
		filePath: filePath,
		queue:    make(chan data)}

	go heap.handle()

	return heap
}

func (heap *Heap) handle() {
	for one := range heap.queue {
		heap.RLock()
		heap.data[one.key] = data{
			value:     one.value,
			timestamp: one.timestamp}
		heap.RUnlock()

		heap.append(one)
	}
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
	heap.queue <- data{
		key:       key,
		value:     value,
		timestamp: time.Now().Unix() + ttl}
}

func (heap *Heap) Get(key string) string {
	one, ok := heap.data[key]

	if ok {
		if one.timestamp < time.Now().Unix() {
			delete(heap.data, key)

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

func (heap *Heap) Save() error {
	os.Remove(heap.filePath)

	for _, one := range heap.data {
		heap.append(one)
	}

	return nil
}

func (heap *Heap) Restore() error {
	_, err := os.Stat(heap.filePath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(heap.filePath, os.O_RDONLY, 0777)
	if err != nil {
		return err
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

			ttl, err := strconv.ParseInt(slices[2], 10, 64)
			if err != nil {
				return err
			}

			heap.data[key] = data{
				value:     value,
				timestamp: ttl}
		}
	}

	return nil
}
