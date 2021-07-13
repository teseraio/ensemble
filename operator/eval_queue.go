package operator

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/teseraio/ensemble/operator/proto"
)

type EvalQueue struct {
	heap     taskQueueImpl
	lock     sync.Mutex
	items    map[string]*evalTask
	updateCh chan struct{}
	pending  map[string][]*proto.Evaluation
}

func NewEvalQueue() *EvalQueue {
	return &EvalQueue{
		heap:     taskQueueImpl{},
		items:    map[string]*evalTask{},
		updateCh: make(chan struct{}),
		pending:  map[string][]*proto.Evaluation{},
	}
}

func (e *EvalQueue) Add(eval *proto.Evaluation) {
	e.lock.Lock()
	defer e.lock.Unlock()

	found := false
	for _, i := range e.items {
		if i.eval.DeploymentID == eval.DeploymentID {
			found = true
			break
		}
	}

	if found {
		// there is already a task for the same cluster, append
		// this evaluation to the pending map
		if _, ok := e.pending[eval.DeploymentID]; !ok {
			e.pending = map[string][]*proto.Evaluation{}
		}
		e.pending[eval.DeploymentID] = append(e.pending[eval.DeploymentID], eval)
	} else {
		e.addImpl(eval)
	}
}

func (e *EvalQueue) addImpl(eval *proto.Evaluation) {
	tt := &evalTask{
		eval:      eval,
		timestamp: time.Now(),
		ready:     true,
	}

	e.items[eval.Id] = tt
	heap.Push(&e.heap, tt)

	select {
	case e.updateCh <- struct{}{}:
	default:
	}
}

func (e *EvalQueue) popImpl() *proto.Evaluation {
	e.lock.Lock()
	if len(e.heap) != 0 && e.heap[0].ready {
		// pop the first value and remove it from the heap
		tt := e.heap[0]
		tt.ready = false
		heap.Fix(&e.heap, tt.index)
		e.lock.Unlock()

		return tt.eval
	}
	e.lock.Unlock()
	return nil
}

func (e *EvalQueue) Pop(ctx context.Context) *proto.Evaluation {
POP:
	tt := e.popImpl()
	if tt != nil {
		return tt
	}

	select {
	case <-e.updateCh:
		goto POP
	case <-ctx.Done():
		return nil
	}
}

func (e *EvalQueue) Finalize(id string) bool {
	e.lock.Lock()
	defer e.lock.Unlock()

	i, ok := e.items[id]
	if !ok {
		return false
	}

	// remove the element from the heap
	heap.Remove(&e.heap, i.index)
	delete(e.items, id)

	// check if there is a pending eval
	pending, ok := e.pending[i.eval.DeploymentID]
	if ok {
		var nextTask *proto.Evaluation
		nextTask, pending = pending[0], pending[1:]
		if len(pending) == 0 {
			delete(e.pending, i.eval.DeploymentID)
		} else {
			e.pending[i.eval.DeploymentID] = pending
		}
		e.addImpl(nextTask)
	}
	return true
}

type evalTask struct {
	eval      *proto.Evaluation
	index     int
	timestamp time.Time
	ready     bool
}

type taskQueueImpl []*evalTask

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
	item := x.(*evalTask)
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
