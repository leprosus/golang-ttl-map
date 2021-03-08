package ttl_map

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	heap := New()

	switch fmt.Sprintf("%T", heap) {
	case "*ttl_map.Heap":
	default:
		t.Error("New return Heap var")
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	filePath := "./set_tmp"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("key", "value", 60)

	time.Sleep(500 * time.Microsecond)

	_, err := os.Stat(filePath)
	if err != nil {
		t.Error("After Set procedure a file ", filePath, "wasn't saved")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	filePath := "./get_tmp"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("key", "value", 1)

	value, ok := heap.Get("key")
	if !ok {
		t.Error("Can't get expected value by `key`")
	}

	if value.(string) != "value" {
		t.Error("Get returns set value")
		return
	}

	time.Sleep(time.Second)

	value, ok = heap.Get("key")
	if ok {
		t.Error("Get unexpected value by `key`")
	}

	if value != nil {
		t.Error("Get returns empty value for key with non-valid ttl")
		return
	}
}

func TestDel(t *testing.T) {
	t.Parallel()

	filePath := "./del_tmp"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("key", "value", 1)

	heap.Del("key")

	value, ok := heap.Get("key")
	if ok {
		t.Error("Get unexpected value by `key`")
	}

	if value != nil {
		t.Error("Get returns empty value after deleting")
	}
}

func TestSave(t *testing.T) {
	t.Parallel()

	filePath := "./save_tmp"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("key", "value", 1)

	err := heap.Save()
	if err != nil {
		t.Error(err)
	}

	heap.Wait()
}

func TestRestore(t *testing.T) {
	t.Parallel()

	filePath := "./restore_tmp"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()

		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("key", "value", 60)
	heap.Set("key1", "value", 60)

	err := heap.Save()
	if err != nil {
		t.Error(err)
	}

	heap.Wait()

	heap = New()
	heap.Path(filePath)
	err = heap.Restore()
	if err != nil {
		t.Error(err)
	}

	value, ok := heap.Get("key")
	if !ok {
		t.Error("Can't get expected value by `key`")
	}

	if value.(string) != "value" {
		t.Error("Restore restores data from file")
		return
	}
}

func TestSupport(t *testing.T) {
	t.Parallel()

	filePath := "./support"
	heap := New()
	heap.Path(filePath)

	defer func() {
		heap.Wait()
		err := os.Remove(filePath)
		if err != nil {
			t.Error(err)
		}
	}()

	heap.Set("object", map[string]string{"key": "value"}, 60)

	heap.Support(map[string]string{})

	err := heap.Save()
	if err != nil {
		t.Error(err)
	}

	heap.Wait()

	heap = New()
	heap.Path(filePath)
	err = heap.Restore()
	if err != nil {
		t.Error(err)
	}

	val, ok := heap.Get("object")
	if !ok {
		t.Error("Can't get map structure")
	}

	var obj = val.(map[string]string)

	var objVal string

	objVal, ok = obj["key"]
	if !ok {
		t.Error("Can't get value from map structure")
	}

	if objVal != "value" {
		t.Error("Object doesn't contain expected value by the key")
	}
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	filePath := "./goroutin_tmp"
	heap := New()
	heap.Path(filePath)

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
	err := os.Remove(filePath)
	if err != nil {
		t.Error(err)
	}
}
