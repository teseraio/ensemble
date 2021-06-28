package boltdb

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
)

type task2 struct {
	*proto.Task
	clusterID string

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

func (t *taskQueue2) add(task *proto.Task) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt := &task2{
		clusterID: task.DeploymentID,
		Task:      task,
		ready:     true,
	}

	t.items[tt.clusterID] = tt
	heap.Push(&t.heap, tt)

	select {
	case t.updateCh <- struct{}{}:
	default:
	}
}

func (t *taskQueue2) popImpl() *task2 {
	t.lock.Lock()
	if len(t.heap) != 0 && t.heap[0].ready {
		// pop the first value and remove it from the heap
		tt := t.heap[0]
		tt.ready = false
		heap.Fix(&t.heap, tt.index)
		t.lock.Unlock()

		return tt
	}
	t.lock.Unlock()
	return nil
}

func (t *taskQueue2) pop(ctx context.Context) *task2 {
POP:
	tt := t.popImpl()
	if tt != nil {
		return tt
	}

	select {
	case <-t.updateCh:
		goto POP
	case <-ctx.Done():
		return nil
	}
}

func (t *taskQueue2) finalize(clusterID string) (*task2, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	item, ok := t.items[clusterID]
	if !ok {
		return nil, false
	}

	// remove the element from the heap
	heap.Remove(&t.heap, item.index)
	delete(t.items, item.clusterID)

	return item, true
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
