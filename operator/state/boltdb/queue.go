package boltdb

import (
	"container/heap"
	"context"
	"fmt"
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

func (t *taskQueue2) get(id string) (*task2, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	tt, ok := t.items[id]
	return tt, ok
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

	fmt.Println(t.items)
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

// -------------------------

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

func (t *taskQueue) finalize(clusterID string) (*task, bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	var item *task
	for _, i := range t.items {
		if i.clusterID == clusterID {
			if i.ready {
				return nil, false
			}
			item = i
		}
	}
	if item == nil {
		return nil, false
	}

	// remove the element from the heap
	heap.Remove(&t.heap, item.index)
	delete(t.items, item.Id)

	// check if there is a pending eval
	pending, ok := t.pending[clusterID]
	if ok {
		var nextTask *proto.Component
		nextTask, pending = pending[0], pending[1:]
		if len(pending) == 0 {
			delete(t.pending, clusterID)
		} else {
			t.pending[clusterID] = pending
		}
		t.addImpl(clusterID, nextTask)
	}
	return item, true
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
