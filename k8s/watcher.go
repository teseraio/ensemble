package k8s

import (
	"bufio"
	"container/heap"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/mapstructure"
)

type itemObj interface {
	ResourceVersion() string
	Name() string
}

type listResponse struct {
	Items    []interface{}
	Metadata struct {
		Continue        string
		ResourceVersion string
	}
}

type pagerClient interface {
	Get(path string) (*listResponse, error)
}

type pager struct {
	path   string
	client pagerClient
	limit  int

	// results
	items           []interface{}
	resourceVersion string
}

func (p *pager) run(stopCh chan struct{}) error {
	resourceVersion := p.resourceVersion

	limit := p.limit
	if limit == 0 {
		limit = 20
	}

	cont := ""

	for {
		path := p.path
		if !strings.Contains(path, "?") {
			path = path + "?"
		} else {
			path = path + "&"
		}
		if resourceVersion != "" && cont == "" {
			// continue and resourceVersion cannot be set at the same time
			path += "resourceVersion=" + resourceVersion + "&"
		}
		if cont != "" {
			path += "continue=" + cont + "&"
		}
		if limit != 0 {
			path += "limit=" + strconv.Itoa(limit)
		}

		result, err := p.client.Get(path)
		if err != nil {
			return err
		}
		p.items = append(p.items, result.Items...)
		if result.Metadata.Continue == "" {
			resourceVersion = result.Metadata.ResourceVersion
			break
		}
		cont = result.Metadata.Continue
	}

	p.resourceVersion = resourceVersion
	return nil
}

type WatchEntry struct {
	item itemObj
	typ  string

	// internal fields for the sort heap
	index     int
	timestamp time.Time
}

type Watcher struct {
	logger                hclog.Logger
	client                *KubeClient
	stopCh                chan struct{}
	store                 *store
	path                  string
	list                  bool
	limit                 int
	obj                   itemObj
	latestResourceVersion string
}

func NewWatcher(logger hclog.Logger, client *KubeClient, path string, obj itemObj) (*Watcher, error) {
	w := &Watcher{
		logger: logger,
		client: client,
		store:  newStore(),
		obj:    obj,
		path:   path,
	}

	// validate that we can get the items
	if _, err := w.client.GetFull(path, nil); err != nil {
		return nil, err
	}
	// validate that we can watch the items
	if _, err := w.client.Watch(path); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) decodeObj(item interface{}) (itemObj, error) {
	obj := reflect.New(reflect.TypeOf(w.obj).Elem()).Interface()
	if err := mapstructure.Decode(item, obj); err != nil {
		return nil, err
	}
	return obj.(itemObj), nil
}

func (w *Watcher) watchImpl(resourceVersion string, handler func(typ string, item itemObj) error) error {
	path := w.path + "?watch=true"
	if resourceVersion != "" {
		path += "&resourceVersion=" + resourceVersion
	}

	resp, err := w.client.Watch(path)
	if err != nil {
		return err
	}

	buffer := bufio.NewReader(resp.Body)
	for {
		res, err := buffer.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}

		if err := isError(res); err != nil {
			return err
		}

		var evnt WatchEvent
		if err := json.Unmarshal(res, &evnt); err != nil {
			w.logger.Error("failed to decode watch event", "err", err)
			continue
		}
		obj, err := w.decodeObj(evnt.Object)
		if err != nil {
			w.logger.Error("failed to decode custom event", "err", err)
			continue
		}
		if err := handler(evnt.Type, obj); err != nil {
			return err
		}
	}
}

func (w *Watcher) Get(path string) (*listResponse, error) {
	data, err := w.client.GetFull(path, nil)
	if err != nil {
		return nil, err
	}
	result := &listResponse{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (w *Watcher) WithLimit(l int) *Watcher {
	w.limit = l
	return w
}

func (w *Watcher) WithList(list bool) *Watcher {
	w.list = list
	return w
}

func (w *Watcher) Run(stopCh chan struct{}) {
	go w.runWithBackoff(stopCh)
}

func (w *Watcher) runWithBackoff(stopCh chan struct{}) {
	for {
		err := w.runImpl()
		if err != nil {
			w.logger.Error("failed to watch", "err", err)
		}

		// TODO: Use exponential backoff
		select {
		case <-time.After(2 * time.Second):
		case <-stopCh:
			return
		}
	}
}

func (w *Watcher) listImpl(resourceVersion string) (*pager, error) {
	pager := &pager{
		path:            w.path,
		client:          w,
		limit:           w.limit,
		resourceVersion: resourceVersion,
	}
	err := pager.run(w.stopCh)
	return pager, err
}

func isExpired(err error) bool {
	return err == errExpired
}

func (w *Watcher) runImpl() error {
	resourceVersion := w.latestResourceVersion
	if resourceVersion == "" {
		// start from the beginning
		resourceVersion = "0"
	}

	defer func() {
		w.latestResourceVersion = resourceVersion
	}()

	// try to list first with the latest known resourceversion
	pager, err := w.listImpl(resourceVersion)
	if err != nil {
		if isExpired(err) {
			// the resource version is not available, retry againg
			// but this time from the latest known value
			pager, err = w.listImpl("")
		}
	}
	if err != nil {
		return err
	}

	// decode and insert all the objects
	if w.list {
		objs := []itemObj{}
		for _, i := range pager.items {
			item, err := w.decodeObj(i)
			if err != nil {
				return err
			}
			objs = append(objs, item)
		}
		for _, obj := range objs {
			w.store.add("", obj)
		}
	}

	resourceVersion = pager.resourceVersion

	// watch
	err = w.watchImpl(resourceVersion, func(typ string, item itemObj) error {
		w.store.add(typ, item)
		resourceVersion = item.ResourceVersion()
		return nil
	})
	return err
}

func (w *Watcher) ForEach(handler func(task *WatchEntry, i interface{})) {
	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		<-w.stopCh
		cancelFn()
	}()

	go func() {
		for {
			task := w.store.pop(ctx)
			if ctx.Err() != nil {
				return
			}
			handler(task, task.item)
		}
	}()
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
	id := i.Name()
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
		delete(s.items, tt.item.Name())
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
	return t[i].item.ResourceVersion() < t[j].item.ResourceVersion()
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
