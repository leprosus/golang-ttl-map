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
	value string
	ttl   int64
}

type Heap struct {
	data  map[string]data
	mutex *sync.Mutex
}

func New() Heap {
	heap := Heap{
		data:  map[string]data{},
		mutex: &sync.Mutex{}}

	return heap
}

func (heap *Heap) Set(key string, value string, ttl int64) {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	heap.data[key] = data{
		value: value,
		ttl:   time.Now().Unix() + ttl}
}

func (heap *Heap) Get(key string) string {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if data, ok := heap.data[key]; ok {
		if data.ttl < time.Now().Unix() {
			heap.Del(key)

			return ""
		} else {
			return data.value
		}
	} else {
		return ""
	}
}

func (heap *Heap) Del(key string) {
	delete(heap.data, key)
}

func (heap *Heap) Save(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	writer := bufio.NewWriter(file)
	for key, data := range heap.data {
		line := fmt.Sprintf("%s\t%s\t%d\n", key, data.value, data.ttl)

		if num, err := writer.WriteString(line); err == nil && num < len(line) {
			return io.ErrShortWrite
		}
		writer.Flush()
	}

	return nil
}

func (heap *Heap) Restore(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	heap = &Heap{
		data:  map[string]data{},
		mutex: &sync.Mutex{}}

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
				value: value,
				ttl:   time.Now().Unix() + ttl}
		}
	}

	return nil
}
