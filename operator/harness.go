package operator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/lib/uuid"
	"github.com/teseraio/ensemble/operator/proto"
)

type Harness struct {
	t          assert.TestingT
	Deployment *proto.Deployment
	Handler    Handler
	Scheduler  Scheduler
	Component  *proto.Component
}

func NewHarness(t assert.TestingT) *Harness {
	return &Harness{
		t:          t,
		Deployment: &proto.Deployment{},
	}
}

func (h *Harness) AddComponent(c *proto.Component) {
	if c.Id == "" {
		c.Id = uuid.UUID()
	}
	h.Component = c
	h.Deployment.CompId = c.Id
}

func (h *Harness) ApplyDep(plan *proto.Plan, callback func(i *proto.Instance)) *proto.Deployment {
	dep := h.Deployment.Copy()

	for _, n := range plan.NodeUpdate {
		callback(n)

		exists := -1
		for indx, i := range dep.Instances {
			if i.ID == n.ID {
				exists = indx
				break
			}
		}
		if exists == -1 {
			dep.Instances = append(dep.Instances, n)
		} else {
			dep.Instances[exists] = n
		}
	}

	h.Deployment = dep
	return dep
}

func (h *Harness) GetComponentByID(dep, id string, sequence int64) (*proto.Component, error) {
	return h.Component, nil
}

func (h *Harness) LoadDeployment(id string) (*proto.Deployment, error) {
	if h.Deployment == nil {
		h.Deployment = &proto.Deployment{}
	}
	return h.Deployment, nil
}

func (h *Harness) GetHandler(id string) (Handler, error) {
	return h.Handler, nil
}

func (h *Harness) Eval() *proto.Plan {
	plan, err := h.Scheduler.Process(&proto.Evaluation{})
	assert.NoError(h.t, err)
	return plan
}

type HarnessExpect struct {
	Status string
	Nodes  []*HarnessExpectInstance
}

type HarnessExpectInstance struct {
	Name   string
	KV     map[string]string
	Status proto.Instance_Status
	Spec   *proto.NodeSpec
}

func (h *Harness) Expect(plan *proto.Plan, expect *HarnessExpect) {
	if expect.Status != "" {
		if plan.Status != expect.Status {
			assert.Equal(h.t, plan.Status, expect.Status)
		}
	}

	applyTmpl := func() func(v string) string {
		obj := map[string]interface{}{}

		instances := []*proto.Instance{}
		instances = append(instances, h.Deployment.Instances...)
		instances = append(instances, plan.NodeUpdate...)

		for _, i := range instances {
			indx, err := proto.ParseIndex(i.Name)
			if err != nil {
				assert.Error(h.t, err)
			}
			if i.Group.Type == "" {
				obj[fmt.Sprintf("Node_%d", indx)] = i.Name
			} else {
				obj[fmt.Sprintf("Node_%s_%d", i.Group.Type, indx)] = i.Name
			}
		}
		return func(v string) string {
			t, err := template.New("").Parse(v)
			if err != nil {
				assert.Error(h.t, err)
			}
			buf1 := new(bytes.Buffer)
			if err = t.Execute(buf1, obj); err != nil {
				assert.Error(h.t, err)
			}
			return buf1.String()
		}
	}

	tmpl := applyTmpl()

	assert.Equal(h.t, len(plan.NodeUpdate), len(expect.Nodes))

	// compare nodes
	for i := 0; i < len(plan.NodeUpdate); i++ {
		node := plan.NodeUpdate[i]
		expect := expect.Nodes[i]

		for k, v := range expect.KV {
			assert.Equal(h.t, node.KV[k], v)
		}
		if expect.Name != "" {
			assert.Equal(h.t, node.Name, tmpl(expect.Name))
		}

		if expect.Spec != nil {
			for k, v := range expect.Spec.Env {
				assert.Equal(h.t, node.Spec.Env[k], tmpl(v))
			}
			{
				// args
				assert.Equal(h.t, len(expect.Spec.Args), len(node.Spec.Args))
				for indx, argExp := range expect.Spec.Args {
					assert.Equal(h.t, node.Spec.Args[indx], tmpl(argExp))
				}
			}
		}
	}
}
