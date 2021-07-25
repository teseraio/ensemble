package zookeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test4lw_StatUnmarshal(t *testing.T) {
	raw := `Zookeeper version: 3.6.2--803c7f1a12f85978cb049af5e4ef23bd8b688715, built on 09/04/2020 12:44 GMT
Clients:
	/172.23.0.1:44810[0](queued=0,recved=1,sent=0)

Latency min/avg/max: 0/0.0/0
Received: 2
Sent: 1
Connections: 1
Outstanding: 0
Zxid: 0x1000003e9
Mode: follower
Node count: 1005`

	s := &Stat{}
	assert.NoError(t, s.Unmarshal(raw))

	assert.Equal(t, s.Mode, "follower")
	assert.Equal(t, s.NodeCount, int64(1005))
	assert.Equal(t, s.Epoch, int64(1))
	assert.Equal(t, s.Counter, int64(1001))
}

func Test4l2_ConfFileUnmarshal(t *testing.T) {
	raw := `clientPort=2181
secureClientPort=-1
dataDir=/data/version-2
dataDirSize=0
dataLogDir=/datalog/version-2
dataLogSize=1266
tickTime=2000
maxClientCnxns=60
minSessionTimeout=4000
maxSessionTimeout=40000
clientPortListenBacklog=-1
serverId=3
initLimit=5
syncLimit=2
electionAlg=3
electionPort=3888
quorumPort=2888
peerType=0
membership: 
server.1=342330ff-1.A:2888:3888:participant;0.0.0.0:2181
server.2=c990df62-2.A:2888:3888:participant;0.0.0.0:2181
server.3=0.0.0.0:2888:3888:participant;0.0.0.0:2181
version=0`

	c := &ConfFile{}
	assert.NoError(t, c.Unmarshal(raw))

	assert.Equal(t, c.Get("tickTime"), "2000")
}
