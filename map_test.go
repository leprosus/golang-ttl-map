package ttl_map

import (
	"testing"
	"fmt"
	"os"
	"time"
	"io/ioutil"
	"strings"
)

func TestNew(t *testing.T) {
	t.Parallel()

	filePath := "./new_tmp.tsv"
	heap := New(filePath)

	switch fmt.Sprintf("%T", heap) {
	case "ttl_map.Heap":
	default:
		t.Error("Expected New return Heap var")
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	filePath := "./set_tmp.tsv"
	heap := New(filePath)

	defer func() {
		time.Sleep(time.Second)
		os.Remove(filePath)
	}()

	heap.Set("key", "value", 60)

	time.Sleep(500 * time.Microsecond)

	_, err := os.Stat(filePath)
	if err != nil {
		t.Error("Expected after Set file ", filePath, "will be saved")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	filePath := "./get_tmp.tsv"
	heap := New(filePath)

	defer func() {
		time.Sleep(time.Second)
		os.Remove(filePath)
	}()

	heap.Set("key", "value", 1)

	value := heap.Get("key")

	if value != "value" {
		t.Error("Expected Get returns set value")
		return
	}

	time.Sleep(1 * time.Second)

	value = heap.Get("key")

	if value != "" {
		t.Error("Expected Get returns empty value for key with non-valid ttl")
		return
	}
}

func TestDel(t *testing.T) {
	t.Parallel()

	filePath := "./del_tmp.tsv"
	heap := New(filePath)

	defer func() {
		time.Sleep(time.Second)
		os.Remove(filePath)
	}()

	heap.Set("key", "value", 1)

	heap.Del("key")

	value := heap.Get("key")

	if value != "" {
		t.Error("Expected Get returns empty value after deleting")
	}
}

func TestSave(t *testing.T) {
	t.Parallel()

	filePath := "./save_tmp.tsv"
	heap := New(filePath)

	defer func() {
		time.Sleep(time.Second)
		os.Remove(filePath)
	}()

	now := time.Now().Unix()

	heap.Set("key", "value", 1)

	heap.Save()
	time.Sleep(1 * time.Second)

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Error("Expected Save saves data to file")
		return
	}

	line := string(bytes)
	testStr := fmt.Sprintf("key\tvalue\t%d", now)
	testStr = testStr[:len(testStr)-1]

	if !strings.Contains(line, testStr) {
		t.Error("Expected Save saves target data")
		return
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	filePath := "./restore_tmp.tsv"
	heap := New(filePath)

	defer func() {
		time.Sleep(time.Second)
		os.Remove(filePath)
	}()

	heap.Set("key", "value", 60)

	heap.Save()
	time.Sleep(1 * time.Second)

	heap = New(filePath)
	heap.Restore()

	if heap.Get("key") != "value" {
		t.Error("Expected Restore restores data from file")
		return
	}
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	filePath := "./goroutin_tmp.tsv"
	heap := New(filePath)

	go func() {
		for i := 0; i < 1000000; i++ {
			heap.Set(fmt.Sprintf("key_%d", i), "value", 60)
		}
	}()

	go func() {
		for i := 0; i < 1000000; i++ {
			heap.Get(fmt.Sprintf("key_%d", i))
		}
	}()

	time.Sleep(15 * time.Second)
	os.Remove(filePath)
}
