package boltdb

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
)

type task struct {
	*proto.ComponentTask

	// internal fields for the sort heap
	ready     bool
	index     int
	timestamp time.Time
}

type taskQueue struct {
	heap     taskQueueImpl
	lock     sync.Mutex
	items    map[string]*task
	updateCh chan struct{}
}

func newTaskQueue() *taskQueue {
	return &taskQueue{
		heap:     taskQueueImpl{},
		items:    map[string]*task{},
		updateCh: make(chan struct{}),
	}
}

func (t *taskQueue) existsByName(name string) bool {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, i := range t.items {
		if i.New.Name == name {
			return true
		}
	}
	return false
}

func (t *taskQueue) get(id string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt, ok := t.items[id]
	return tt, ok
}

func (t *taskQueue) add(new, old *proto.Component) {
	tt := &task{
		ComponentTask: &proto.ComponentTask{
			Old: old,
			New: new,
		},
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	tt.ready = true

	t.items[new.Id] = tt
	heap.Push(&t.heap, tt)

	select {
	case t.updateCh <- struct{}{}:
	default:
	}
}

func (t *taskQueue) pop(ctx context.Context) *task {
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

func (t *taskQueue) finalize(id string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	i, ok := t.items[id]
	if ok {
		heap.Remove(&t.heap, i.index)
		delete(t.items, id)
	}
	return i, ok
}

type taskQueueImpl []*task

func (t taskQueueImpl) Len() int { return len(t) }

func (t taskQueueImpl) Less(i, j int) bool {
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

func (t taskQueueImpl) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *taskQueueImpl) Push(x interface{}) {
	n := len(*t)
	item := x.(*task)
	item.index = n
	*t = append(*t, item)
}

func (t *taskQueueImpl) Pop() interface{} {
	old := *t
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*t = old[0 : n-1]
	return item
}
