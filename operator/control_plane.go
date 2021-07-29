package operator

import (
	"sync"

	"github.com/teseraio/ensemble/operator/proto"
)

type InstanceUpdate struct {
	InstanceID string
}

type ControlPlane interface {
	UpsertInstance(*proto.Instance) error
	GetInstance(instanceID string) (*proto.Instance, error)
	SubscribeInstanceUpdates() <-chan *InstanceUpdate
}

type InmemControlPlane struct {
	lock      sync.Mutex
	instances map[string]*proto.Instance
	subs      []chan *InstanceUpdate
}

func (i *InmemControlPlane) UpsertInstance(ii *proto.Instance) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.instances == nil {
		i.instances = map[string]*proto.Instance{}
	}
	i.instances[ii.ID] = ii
	update := &InstanceUpdate{
		InstanceID: ii.ID,
	}
	for _, ch := range i.subs {
		select {
		case ch <- update:
		default:
		}
	}
	return nil
}

func (i *InmemControlPlane) GetInstance(InstanceID string) (*proto.Instance, error) {
	ii, ok := i.instances[InstanceID]
	if !ok {
		return nil, nil
	}
	return ii.Copy(), nil
}

func (i *InmemControlPlane) SubscribeInstanceUpdates() <-chan *InstanceUpdate {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.subs == nil {
		i.subs = []chan *InstanceUpdate{}
	}
	ch := make(chan *InstanceUpdate, 10)
	i.subs = append(i.subs, ch)

	return ch
}
