package zookeeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"github.com/teseraio/ensemble/operator"
	"github.com/teseraio/ensemble/operator/proto"
	"github.com/teseraio/ensemble/schema"
	"github.com/teseraio/ensemble/testutil"
)

func TestBootstrap(t *testing.T) {
	h := operator.NewHarness(t)
	h.Handler = Factory()
	h.Scheduler = operator.NewScheduler(h)

	h.AddComponent(&proto.Component{
		Id: "a1",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count:  3,
					Params: schema.MapToSpec(nil),
				},
			},
		}),
	})

	plan := h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "1",
						"ZOO_SERVERS": "server.1=0.0.0.0:2888:3888;2181 server.2={{.Node_2}}:2888:3888;2181 server.3={{.Node_3}}:2888:3888;2181",
					},
				},
			},
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "2",
						"ZOO_SERVERS": "server.1={{.Node_1}}:2888:3888;2181 server.2=0.0.0.0:2888:3888;2181 server.3={{.Node_3}}:2888:3888;2181",
					},
				},
			},
			{
				Spec: &proto.NodeSpec{
					Env: map[string]string{
						"ZOO_MY_ID":   "3",
						"ZOO_SERVERS": "server.1={{.Node_1}}:2888:3888;2181 server.2={{.Node_2}}:2888:3888;2181 server.3=0.0.0.0:2888:3888;2181",
					},
				},
			},
		},
	})

	h.ApplyDep(plan, func(n *proto.Instance) {
		n.Status = proto.Instance_RUNNING
		n.Healthy = true
	})

	// it should be done once all nodes are running
	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Status: "done",
	})

	// Update tick time
	h.AddComponent(&proto.Component{
		Id:       "a2",
		Sequence: 1,
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(map[string]interface{}{
						"tickTime": "3000",
					}),
				},
			},
		}),
	})

	plan = h.Eval()
	h.Expect(plan, &operator.HarnessExpect{
		Nodes: []*operator.HarnessExpectInstance{
			{
				Name:   "{{.Node_1}}",
				Status: proto.Instance_TAINTED,
			},
			{
				Name:   "{{.Node_2}}",
				Status: proto.Instance_TAINTED,
			},
		},
	})
}

func testCorrectness(t *testing.T, srv *testutil.TestServer, name string) {
	dep := srv.GetDeployment(name)

	leaderFound := false

	var stat *Stat
	for _, i := range dep.Instances {
		elem, err := dialStat(i.Ip)
		assert.NoError(t, err)

		fmt.Println("-- elem --")
		fmt.Println(elem)

		if stat == nil {
			stat = elem
		} else {
			assert.Equal(t, stat.Counter, elem.Counter)
			assert.Equal(t, stat.Epoch, elem.Epoch)
			assert.Equal(t, stat.NodeCount, elem.NodeCount)
		}
		if elem.Mode == leader {
			if leaderFound {
				t.Fatal("multiple leaders")
			}
			leaderFound = true
		}
	}
}

func populateKV(srv *testutil.TestServer, num int) error {
	dep := srv.GetDeployment("A")

	c, _, err := zk.Connect([]string{dep.Instances[1].Ip}, time.Second)
	if err != nil {
		return err
	}

	for i := 0; i < num; i++ {
		_, err := c.Create(fmt.Sprintf("/item%d", i), []byte{0x1, 0x2}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	}
	return nil
}

func TestZookeeper_Initial(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
				},
			},
		}),
	})

	srv.WaitForTask(uuid)
	assert.NoError(t, populateKV(srv, 1000))

	testCorrectness(t, srv, "A")
}

func TestZookeeper_FailedLeader(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
				},
			},
		}),
	})

	srv.WaitForComplete("A")
	assert.NoError(t, populateKV(srv, 1000))

	leader := srv.LoadInstance("A", func(i *proto.Instance) bool {
		stat, err := dialStat(i.Ip)
		assert.NoError(t, err)
		return stat.Mode == leader
	})
	srv.Remove(leader.ID)
	srv.WaitForComplete("A")

	testCorrectness(t, srv, "A")
}

func TestZookeeper_FailedFollower(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
				},
			},
		}),
	})

	srv.WaitForComplete("A")
	assert.NoError(t, populateKV(srv, 1000))

	follower := srv.LoadInstance("A", func(i *proto.Instance) bool {
		stat, err := dialStat(i.Ip)
		assert.NoError(t, err)
		return stat.Mode == follower
	})
	srv.Remove(follower.ID)
	srv.WaitForComplete("A")

	testCorrectness(t, srv, "A")
}

func TestZookeeper_Configuration(t *testing.T) {
	srv := testutil.TestOperator(t, Factory)
	// defer srv.Close()

	uuid := srv.Apply(&proto.Component{
		Name: "A",
		Spec: proto.MustMarshalAny(&proto.ClusterSpec{
			Backend: "Zookeeper",
			Groups: []*proto.ClusterSpec_Group{
				{
					Count: 3,
					Params: schema.MapToSpec(map[string]interface{}{
						"tickTime": 1001,
					}),
				},
			},
		}),
	})

	srv.WaitForTask(uuid)

	ii := srv.LoadInstance("A", first)
	conf, err := dialConf(ii.Ip + ":2181")
	assert.NoError(t, err)
	assert.Equal(t, conf.Get("tickTime"), "1001")
}

func first(i *proto.Instance) bool {
	return true
}
