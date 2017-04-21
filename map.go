package ttl_map

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type data struct {
	value     string
	timestamp int64
}

type Heap struct {
	data     map[string]data
	mutex    *sync.Mutex
	filePath string
}

func New(filePath string) Heap {
	heap := Heap{
		data:     map[string]data{},
		mutex:    &sync.Mutex{},
		filePath: filePath}

	return heap
}

func (heap *Heap) Set(key string, value string, ttl int64) {
	heap.mutex.Lock()

	heap.data[key] = data{
		value:     value,
		timestamp: time.Now().Unix() + ttl}

	heap.mutex.Unlock()
}

func (heap *Heap) Get(key string) string {
	heap.mutex.Lock()

	one, ok := heap.data[key]

	heap.mutex.Unlock()

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
	heap.mutex.Lock()

	delete(heap.data, key)

	heap.mutex.Unlock()
}

func (heap *Heap) Save() error {
	file, err := os.OpenFile(heap.filePath + ".sav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Sync()
	defer file.Close()

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	writer := bufio.NewWriter(file)
	for key, data := range heap.data {
		if data.timestamp < time.Now().Unix() {
			delete(heap.data, key)
			continue
		}

		line := fmt.Sprintf("%s\t%s\t%d\n", key, data.value, data.timestamp)

		if num, err := writer.WriteString(line); err == nil && num < len(line) {
			return io.ErrShortWrite
		}
		writer.Flush()
	}

	os.Rename(heap.filePath + ".sav", heap.filePath)

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

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

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
