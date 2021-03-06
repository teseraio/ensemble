package boltdb

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
)

type task struct {
	*proto.Component
	clusterID string

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
	pending  map[string][]*proto.Component
}

func newTaskQueue() *taskQueue {
	return &taskQueue{
		heap:     taskQueueImpl{},
		items:    map[string]*task{},
		updateCh: make(chan struct{}),
		pending:  map[string][]*proto.Component{},
	}
}

func (t *taskQueue) get(id string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt, ok := t.items[id]
	return tt, ok
}

func (t *taskQueue) add(clusterID string, c *proto.Component) {
	t.lock.Lock()
	defer t.lock.Unlock()

	found := false
	for _, i := range t.items {
		if i.clusterID == clusterID {
			found = true
			break
		}
	}

	if found {
		// there is already a task for the same cluster, append
		// this evaluation to the pending map
		if _, ok := t.pending[clusterID]; !ok {
			t.pending = map[string][]*proto.Component{}
		}
		t.pending[clusterID] = append(t.pending[clusterID], c)
	} else {
		t.addImpl(clusterID, c)
	}
}

func (t *taskQueue) addImpl(clusterID string, c *proto.Component) {
	tt := &task{
		clusterID: clusterID,
		Component: c,
		ready:     true,
	}

	fmt.Println("xxxx")
	fmt.Println(c.Id)

	t.items[c.Id] = tt
	heap.Push(&t.heap, tt)

	select {
	case t.updateCh <- struct{}{}:
	default:
	}
}

func (t *taskQueue) popImpl() *task {
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

func (t *taskQueue) pop(ctx context.Context) *task {
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

func (t *taskQueue) finalize(id string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	i, ok := t.items[id]
	if !ok {
		return nil, false
	}

	// remove the element from the heap
	heap.Remove(&t.heap, i.index)
	delete(t.items, id)

	// check if there is a pending eval
	pending, ok := t.pending[i.clusterID]
	if ok {
		var nextTask *proto.Component
		nextTask, pending = pending[0], pending[1:]
		if len(pending) == 0 {
			delete(t.pending, i.clusterID)
		} else {
			t.pending[i.clusterID] = pending
		}
		t.addImpl(i.clusterID, nextTask)
	}
	return i, true
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
