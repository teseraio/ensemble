package boltdb

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
)

type task2 struct {
	*proto.Evaluation

	// internal fields for the sort heap
	ready     bool
	index     int
	timestamp time.Time
}

type taskQueue2 struct {
	heap     taskQueueImpl2
	lock     sync.Mutex
	items    map[string]*task2
	updateCh chan struct{}
}

func newTaskQueue2() *taskQueue2 {
	return &taskQueue2{
		heap:     taskQueueImpl2{},
		items:    map[string]*task2{},
		updateCh: make(chan struct{}),
	}
}

func (t *taskQueue2) get(id string) (*task2, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt, ok := t.items[id]
	return tt, ok
}

func (t *taskQueue2) add(eval *proto.Evaluation) {
	tt := &task2{
		Evaluation: eval,
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	tt.ready = true

	t.items[eval.Id] = tt
	heap.Push(&t.heap, tt)

	select {
	case t.updateCh <- struct{}{}:
	default:
	}
}

func (t *taskQueue2) pop(ctx context.Context) *task2 {
POP:
	t.lock.Lock()
	if len(t.heap) != 0 && t.heap[0].ready {
		// pop the first value
		tt := t.heap[0]
		tt.ready = false
		heap.Fix(&t.heap, tt.index)
		t.lock.Unlock()

		return tt
	}
	t.lock.Unlock()

	select {
	case <-t.updateCh:
		goto POP
	case <-ctx.Done():
		return nil
	}
}

// TODO: REMOVE. Delete the evals on pop
func (t *taskQueue2) finalize(id string) (*task2, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	i, ok := t.items[id]
	if ok {
		heap.Remove(&t.heap, i.index)
		delete(t.items, id)
	}
	return i, ok
}

type taskQueueImpl2 []*task2

func (t taskQueueImpl2) Len() int { return len(t) }

func (t taskQueueImpl2) Less(i, j int) bool {
	iNoReady, jNoReady := !t[i].ready, !t[j].ready
	if iNoReady && jNoReady {
		return false
	} else if iNoReady {
		return false
	} else if jNoReady {
		return true
	}
	return t[i].timestamp.Before(t[j].timestamp)
}

func (t taskQueueImpl2) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *taskQueueImpl2) Push(x interface{}) {
	n := len(*t)
	item := x.(*task2)
	item.index = n
	*t = append(*t, item)
}

func (t *taskQueueImpl2) Pop() interface{} {
	old := *t
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*t = old[0 : n-1]
	return item
}
