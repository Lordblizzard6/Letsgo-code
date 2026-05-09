package tools

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestTaskStore(t *testing.T) *TaskStore {
	t.Helper()
	return &TaskStore{
		Tasks: make(map[string]*Task),
		path:  filepath.Join(t.TempDir(), "tasks.json"),
	}
}

func runWithTimeout(t *testing.T, name string, fn func()) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("%s timed out (possible deadlock)", name)
	}
}

func TestTaskStoreCreateConcurrentNoDeadlock(t *testing.T) {
	store := newTestTaskStore(t)

	runWithTimeout(t, "concurrent create", func() {
		var wg sync.WaitGroup
		workers := 40

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				store.create(fmt.Sprintf("task %d", i))
			}(i)
		}

		wg.Wait()
	})
}

func TestTaskStoreUpdateConcurrentNoDeadlock(t *testing.T) {
	store := newTestTaskStore(t)
	task := store.create("seed")

	runWithTimeout(t, "concurrent update", func() {
		var wg sync.WaitGroup
		workers := 40

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, _ = store.update(task.ID, map[string]interface{}{
					"status": "in_progress",
					"result": fmt.Sprintf("result %d", i),
				})
			}(i)
		}

		wg.Wait()
	})
}

func TestTaskStoreDeleteConcurrentNoDeadlock(t *testing.T) {
	store := newTestTaskStore(t)
	ids := make([]string, 0, 40)
	for i := 0; i < 40; i++ {
		ids = append(ids, store.create(fmt.Sprintf("task %d", i)).ID)
	}

	runWithTimeout(t, "concurrent delete", func() {
		var wg sync.WaitGroup
		for _, id := range ids {
			wg.Add(1)
			go func(taskID string) {
				defer wg.Done()
				_ = store.delete(taskID)
			}(id)
		}

		wg.Wait()
	})
}
