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
	sync.WaitGroup

	data     map[string]data
	filePath string
	queue    chan data

	errFn     func(err error)
	errFnInit bool
}

func New(filePath string) *Heap {
	heap := Heap{
		data:     map[string]data{},
		filePath: filePath,
		queue:    make(chan data, 1000),
	}

	go heap.handle()

	return &heap
}

func (h *Heap) handle() {
	var err error
	for one := range h.queue {
		err = h.append(one)
		if err != nil && h.errFnInit {
			h.errFn(err)
		}
	}
}

func (h *Heap) append(one data) (err error) {
	if one.timestamp < time.Now().Unix() {
		return
	}

	var file *os.File
	file, err = os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Sync()
	}()
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)

	line := fmt.Sprintf("%s\t%s\t%d\n", one.key, one.value, one.timestamp)

	var num int
	num, err = writer.WriteString(line)
	if err == nil && num < len(line) {
		return
	}

	err = writer.Flush()

	return
}

func (h *Heap) Error(fn func(err error)) {
	h.errFn = fn
	h.errFnInit = true
}

func (h *Heap) Set(key string, value string, ttl int64) {
	one := data{
		value:     value,
		timestamp: time.Now().Unix() + ttl}

	h.Lock()
	h.data[key] = one
	h.Unlock()

	one.key = key
	h.queue <- one
}

func (h *Heap) Get(key string) (val string, ok bool) {
	var one data
	h.RLock()
	one, ok = h.data[key]
	h.RUnlock()

	if ok {
		if one.timestamp <= time.Now().Unix() {
			h.Del(key)

			ok = false
		} else {
			val = one.value
		}
	}

	return
}

func (h *Heap) Del(key string) {
	h.Lock()
	delete(h.data, key)
	h.Unlock()
}

func (h *Heap) Save() (err error) {
	var file *os.File
	file, err = os.OpenFile(h.filePath+".sav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}

	writer := bufio.NewWriter(file)

	h.RLock()
	for key, one := range h.data {
		if one.timestamp < time.Now().Unix() {
			continue
		}

		line := fmt.Sprintf("%s\t%s\t%d\n", key, one.value, one.timestamp)

		if num, err := writer.WriteString(line); err == nil && num < len(line) {
			continue
		}
	}
	h.RUnlock()

	err = writer.Flush()
	if err != nil {
		return
	}

	err = file.Close()
	if err != nil {
		return
	}

	err = os.Remove(h.filePath)
	if err != nil {
		return
	}

	err = os.Rename(h.filePath+".sav", h.filePath)

	return
}

func (h *Heap) Restore() (err error) {
	_, err = os.Stat(h.filePath)
	if err != nil {
		return
	}

	var file *os.File
	file, err = os.OpenFile(h.filePath, os.O_RDONLY, 0777)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	h.Lock()
	defer h.Unlock()

	h.data = map[string]data{}

	var timestamp int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		slices := strings.Split(line, "\t")
		if len(slices) == 3 {
			key := slices[0]
			value := slices[1]

			timestamp, err = strconv.ParseInt(slices[2], 10, 64)
			if err != nil {
				return err
			}

			h.data[key] = data{
				value:     value,
				timestamp: timestamp}
		}
	}

	return
}
