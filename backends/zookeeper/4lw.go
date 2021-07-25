package zookeeper

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	follower = "follower"
	leader   = "leader"
)

func dial4lw(server string, command string) ([]byte, error) {
	timeout := 2 * time.Second

	conn, err := net.DialTimeout("tcp", server, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write([]byte(command))
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(timeout))
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type ConfFile struct {
	raw map[string]string
}

func (c *ConfFile) Unmarshal(data string) error {
	c.raw = map[string]string{}

	lines := strings.Split(data, "\n")
	for _, l := range lines {
		indx := strings.Index(l, "=")
		if indx == -1 {
			continue
		}
		key := l[:indx]
		val := strings.TrimSpace(l[indx+1:])
		c.raw[key] = val
	}
	return nil
}

func (c *ConfFile) Get(k string) string {
	return c.raw[k]
}

func (c *ConfFile) GetInt(k string) int {
	k, ok := c.raw[k]
	if !ok {
		return -1
	}
	num, err := strconv.Atoi(k)
	if err != nil {
		return -1
	}
	return num
}

func dialConf(server string) (*ConfFile, error) {
	data, err := dial4lw(server, "conf")
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
	conf := &ConfFile{}
	if err := conf.Unmarshal(string(data)); err != nil {
		return nil, err
	}
	return conf, nil
}

func dialIsReadyForRequests(server string) error {
	data, err := dial4lw(server, "stat")
	if err != nil {
		return err
	}
	str := strings.TrimSpace(string(data))
	if str == "This ZooKeeper instance is not currently serving requests" {
		return fmt.Errorf("not ready")
	}
	return nil
}

type Stat struct {
	NodeCount int64
	Mode      string
	Epoch     int64
	Counter   int64
}

func (s *Stat) Unmarshal(data string) error {
	lines := strings.Split(data, "\n")
	for _, l := range lines {
		indx := strings.Index(l, ":")
		if indx == -1 {
			continue
		}
		title := l[:indx]
		raw := strings.TrimSpace(l[indx+1:])

		switch title {
		case "Node count":
			nodeCount, err := strconv.ParseInt(raw, 0, 64)
			if err != nil {
				return err
			}
			s.NodeCount = nodeCount
		case "Mode":
			switch raw {
			case "follower":
				s.Mode = follower
			case "leader":
				s.Mode = leader
			default:
				return fmt.Errorf("mode not found '%s'", raw)
			}
		case "Zxid":
			zxid, err := strconv.ParseInt(raw, 0, 64)
			if err != nil {
				return err
			}
			s.Epoch = int64(zxid >> 32)
			s.Counter = int64(zxid & 0xFFFFFFFF)
		}
	}
	return nil
}

func dialStat(server string) (*Stat, error) {
	data, err := dial4lw(server+":2181", "stat")
	if err != nil {
		return nil, err
	}
	s := &Stat{}
	if err := s.Unmarshal(string(data)); err != nil {
		return nil, err
	}
	return s, nil
}
