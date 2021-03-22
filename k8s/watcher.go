package k8s

import (
	"bufio"
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

func newWatcher(store *store, client *KubeClient, path string, obj itemObj, list bool) {
	w := &Watcher{
		store:  store,
		client: client,
		path:   path,
		obj:    obj,
		list:   list,
	}
	go w.Start()
}

type Watcher struct {
	store  *store
	client *KubeClient
	path   string
	obj    itemObj
	list   bool
}

func (w *Watcher) Start() {
	w.listImpl()
}

type itemObj interface {
	GetMetadata() *Metadata
}

type ListResponse struct {
	Items    []interface{}
	Metadata ListMetadata
}

type ListMetadata struct {
	Continue        string
	ResourceVersion string
}

type ListOpts struct {
	Continue string
	Limit    int
}

func (w *Watcher) decodeObj(item interface{}) (itemObj, error) {
	obj := reflect.New(reflect.TypeOf(w.obj).Elem()).Interface()
	if err := mapstructure.Decode(item, obj); err != nil {
		return nil, err
	}
	return obj.(itemObj), nil
}

func (w *Watcher) listImpl2() string {
	var resourceVersion string

	// initial list sync
	opts := &ListOpts{
		Limit: 10,
	}
	for {
		path := w.path
		if !strings.Contains(path, "?") {
			path = path + "?"
		} else {
			path = path + "&"
		}
		if opts != nil {
			if opts.Continue != "" {
				path += "continue=" + opts.Continue + "&"
			}
			if opts.Limit != 0 {
				path += "limit=" + strconv.Itoa(opts.Limit)
			}
		}

		data, err := w.client.Get(path)
		if err != nil {
			panic(err)
		}
		result := &ListResponse{}
		if err := json.Unmarshal(data, &result); err != nil {
			fmt.Println(path)
			panic(err)
		}

		if w.list {
			for _, item := range result.Items {
				obj, err := w.decodeObj(item)
				if err != nil {
					panic(err)
				}
				w.store.add("", obj)
			}
		}
		if result.Metadata.Continue == "" {
			resourceVersion = result.Metadata.ResourceVersion
			break
		}
		opts.Continue = result.Metadata.Continue
	}

	return resourceVersion
}

func (w *Watcher) listImpl() {
	resourceVersion := w.listImpl2()

	// initial sync is done, start to watch
	path := w.path + "?watch=true"
	if resourceVersion != "" {
		path += "&resourceVersion=" + resourceVersion
	}
	resp, err := w.client.HTTPReqWithResponse(http.MethodGet, path, nil)
	if err != nil {
		panic(err)
	}

	buffer := bufio.NewReader(resp.Body)
	for {
		res, err := buffer.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				continue
			}
			panic(err)
		}
		var evnt WatchEvent
		if err := json.Unmarshal(res, &evnt); err != nil {
			panic(err)
		}
		obj, err := w.decodeObj(evnt.Object)
		if err != nil {
			panic(err)
		}
		w.store.add(evnt.Type, obj)
	}
}

type WatchEntry struct {
	item itemObj
	typ  string

	// internal fields for the sort heap
	index     int
	timestamp time.Time
}

type store struct {
	heapImpl storeHeapImpl
	lock     sync.Mutex
	items    map[string]*WatchEntry
	updateCh chan struct{}
}

func newStore() *store {
	return &store{
		heapImpl: storeHeapImpl{},
		items:    map[string]*WatchEntry{},
		updateCh: make(chan struct{}),
	}
}

func (s *store) add(typ string, i itemObj) {
	id := i.GetMetadata().Name
	s.lock.Lock()
	defer s.lock.Unlock()

	if ii, ok := s.items[id]; ok {
		// replace
		ii.item = i
		ii.timestamp = time.Now()
		heap.Fix(&s.heapImpl, ii.index)
	} else {
		// push
		tt := &WatchEntry{
			item:      i,
			typ:       typ,
			timestamp: time.Now(),
		}
		s.items[id] = tt
		heap.Push(&s.heapImpl, tt)
	}

	select {
	case s.updateCh <- struct{}{}:
	default:
	}
}

func (s *store) pop(ctx context.Context) *WatchEntry {
POP:
	s.lock.Lock()

	if len(s.heapImpl) != 0 {
		// pop the first value
		tt := heap.Pop(&s.heapImpl).(*WatchEntry)
		delete(s.items, tt.item.GetMetadata().Name)
		s.lock.Unlock()
		return tt
	}
	s.lock.Unlock()

	select {
	case <-s.updateCh:
		goto POP
	case <-ctx.Done():
		return nil
	}
}

type storeHeapImpl []*WatchEntry

func (t storeHeapImpl) Len() int { return len(t) }

func (t storeHeapImpl) Less(i, j int) bool {
	return t[i].timestamp.Before(t[j].timestamp)
}

func (t storeHeapImpl) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
	t[i].index = i
	t[j].index = j
}

func (t *storeHeapImpl) Push(x interface{}) {
	n := len(*t)
	item := x.(*WatchEntry)
	item.index = n
	*t = append(*t, item)
}

func (t *storeHeapImpl) Pop() interface{} {
	old := *t
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*t = old[0 : n-1]
	return item
}
