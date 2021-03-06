package ttl_map

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"io"
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
	dataMx *sync.RWMutex
	fileMx *sync.Mutex
	wg     *sync.WaitGroup

	data     map[string]Data
	filePath string
	queue    chan Data

	errFn     func(err error)
	errFnInit bool
}

func New(filePath string) *Heap {
	heap := Heap{
		dataMx: &sync.RWMutex{},
		fileMx: &sync.Mutex{},
		wg:     &sync.WaitGroup{},

		data:     map[string]Data{},
		filePath: filePath,
		queue:    make(chan Data, 1000),
	}

	go heap.handle()

	return &heap
}

func (h *Heap) handle() {
	var err error
	for data := range h.queue {
		err = h.append(data)
		h.wg.Done()
		if err != nil && h.errFnInit {
			h.errFn(err)
		}
	}
}

func (h *Heap) append(data Data) (err error) {
	h.fileMx.Lock()
	defer h.fileMx.Unlock()

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

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err = enc.Encode(data)
	if err != nil {
		return
	}

	bs := buf.Bytes()
	bs = append(bs, '\n')

	_, err = file.Write(bs)
	if err != nil {
		return
	}

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

	data := Data{
		Key:       key,
		Value:     value,
		Timestamp: time.Now().Unix(),
	}

	if ttl > 0 {
		data.Timestamp += ttl
	} else if ttl < 0 {
		data.Timestamp = -1
	}

	h.dataMx.Lock()
	h.data[key] = data
	h.dataMx.Unlock()

	data.Key = key

	h.wg.Add(1)
	h.queue <- data
}

func (h *Heap) Get(key string) (val interface{}, ok bool) {
	var data Data
	h.dataMx.RLock()
	data, ok = h.data[key]
	h.dataMx.RUnlock()

	if ok {
		if data.Timestamp != -1 && data.Timestamp <= time.Now().Unix() {
			h.Del(key)

			ok = false
		} else {
			val = data.Value
		}
	}

	return
}

func (h *Heap) Del(key string) {
	h.dataMx.RLock()
	_, ok := h.data[key]
	h.dataMx.RUnlock()
	if !ok {
		return
	}

	h.dataMx.Lock()
	delete(h.data, key)
	h.dataMx.Unlock()

	h.wg.Add(1)
	h.queue <- Data{
		Key:       key,
		Timestamp: 0,
	}
}

func (h *Heap) Range(fn func(key string, value interface{}, ttl int64)) {
	h.dataMx.Lock()
	data := h.data
	h.dataMx.Unlock()

	for _, d := range data {
		fn(d.Key, d.Value, d.Timestamp)
	}
}

func (h *Heap) Save() (err error) {
	h.fileMx.Lock()
	defer h.fileMx.Unlock()

	h.wg.Add(1)
	defer h.wg.Done()

	var file *os.File
	file, err = os.OpenFile(h.filePath+".sav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
	}()

	var (
		bs  []byte
		buf bytes.Buffer
	)

	h.dataMx.RLock()
	for _, data := range h.data {
		if data.Timestamp != -1 && data.Timestamp < time.Now().Unix() {
			continue
		}

		buf.Reset()
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(data)
		if err != nil {
			return
		}

		bs = buf.Bytes()
		bs = append(bs, '\n')

		_, err = file.Write(bs)
		if err != nil {
			return
		}
	}
	h.dataMx.RUnlock()

	_ = os.Remove(h.filePath)

	err = os.Rename(h.filePath+".sav", h.filePath)

	return
}

func (h *Heap) Restore() (err error) {
	h.fileMx.Lock()
	defer h.fileMx.Unlock()

	_, err = os.Stat(h.filePath)
	if err != nil {
		return
	}

	var file *os.File
	file, err = os.OpenFile(h.filePath, os.O_RDONLY, 0755)
	if err != nil {
		return
	}
	defer func() {
		_ = file.Sync()
	}()
	defer func() {
		_ = file.Close()
	}()

	reader := bufio.NewReader(file)

	var (
		bs   []byte
		buf  bytes.Buffer
		data Data
		heap = map[string]Data{}
		now  = time.Now().Unix()
	)

	for {
		bs, err = reader.ReadBytes('\n')
		if err == io.EOF {
			err = nil

			break
		}

		if err != nil {
			return
		}

		buf.Reset()
		dec := gob.NewDecoder(&buf)

		bs = bs[:len(bs)-1]
		buf.Write(bs)

		err = dec.Decode(&data)
		if err != nil {
			return
		}

		if data.Timestamp > -1 && data.Timestamp < now {
			continue
		}

		heap[data.Key] = data
	}

	h.dataMx.Lock()
	h.data = heap
	h.dataMx.Unlock()

	return
}

func (h *Heap) Wait() {
	h.wg.Wait()
}
