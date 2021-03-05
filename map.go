package ttl_map

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Data struct {
	Key       string
	Value     interface{}
	Timestamp int64
}

type Heap struct {
	sync.RWMutex
	sync.WaitGroup

	data     map[string]Data
	filePath string
	queue    chan Data

	errFn     func(err error)
	errFnInit bool
}

type snapshot struct {
	Data map[string]Data
}

func New(filePath string) *Heap {
	heap := Heap{
		data:     map[string]Data{},
		filePath: filePath,
		queue:    make(chan Data, 1000),
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

func (h *Heap) append(one Data) (err error) {
	var file *os.File
	file, err = os.OpenFile(h.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
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

	line := fmt.Sprintf("%s\t%s\t%d\n", one.Key, one.Value, one.Timestamp)

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

func (h *Heap) Set(key string, value interface{}, ttl int64) {
	if ttl == 0 {
		return
	}

	one := Data{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().Unix(),
	}

	if ttl > 0 {
		one.Timestamp += ttl
	} else if ttl < 0 {
		one.Timestamp = -1
	}

	h.Lock()
	h.data[key] = one
	h.Unlock()

	one.Key = key
	h.queue <- one
}

func (h *Heap) Get(key string) (val interface{}, ok bool) {
	var one Data
	h.RLock()
	one, ok = h.data[key]
	h.RUnlock()

	if ok {
		if one.Timestamp != -1 && one.Timestamp <= time.Now().Unix() {
			h.Del(key)

			ok = false
		} else {
			val = one.Value
		}
	}

	return
}

func (h *Heap) Del(key string) {
	h.RLock()
	_, ok := h.data[key]
	h.RUnlock()
	if !ok {
		return
	}

	h.Lock()
	delete(h.data, key)
	h.Unlock()

	h.queue <- Data{
		Key:       key,
		Timestamp: 0,
	}
}

func (h *Heap) Range(fn func(key string, value interface{}, ttl int64)) {
	h.Lock()
	data := h.data
	h.Unlock()

	for _, d := range data {
		fn(d.Key, d.Value, d.Timestamp)
	}
}

func (h *Heap) Save() (err error) {
	var snapshot = snapshot{
		Data: map[string]Data{},
	}

	h.RLock()
	for key, one := range h.data {
		if one.Timestamp != -1 && one.Timestamp < time.Now().Unix() {
			continue
		}

		snapshot.Data[key] = one
	}
	h.RUnlock()

	var file *os.File
	file, err = os.OpenFile(h.filePath+".sav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}

	writer := bufio.NewWriter(file)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err = enc.Encode(snapshot)
	if err != nil {
		return
	}

	bs := buf.Bytes()

	_, err = writer.Write(bs)
	if err != nil {
		return
	}

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

	var bs []byte
	bs, err = ioutil.ReadFile(h.filePath)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	buf.Write(bs)

	dec := gob.NewDecoder(&buf)

	var snapshot = snapshot{
		Data: map[string]Data{},
	}

	err = dec.Decode(&snapshot)
	if err != nil {
		return
	}

	h.Lock()
	defer h.Unlock()

	h.data = snapshot.Data

	return
}
