package election

import (
	"github.com/ngaut/log"
	"math/rand"
	"net/rpc"
	"raft/config"
	"raft/rpcs"
	"raft/state"
	"sync"
	"time"
)

const (
	follower = iota
	candidate
	leader
)

// HeartTimerBase must less than TimeoutBase
const TimeoutBase = 3
const HeartTimerBase = 2

var gs = state.GetGlobalState()

type role struct {
	r int
}

func newRole() *role {
	return &role{follower}
}

func (r *role) f() role {
	log.Debug("follower stage")
	timeoutTimer := time.NewTimer(time.Second * TimeoutBase)
	for {
		ri := rand.Intn(10) + TimeoutBase
		timeout := time.Second * time.Duration(ri)

		log.Debugf("random timeout is %+v", timeout)

		select {
		case <-timeoutTimer.C:
			log.Debug("follower timeout, prepare to vote")
			timeoutTimer.Reset(timeout)
			if candidateStage() == leader {
				sendHeartBeatsForever()
			}
		case <-gs.LeaderHeartBeat:
			log.Debugf("get heart beat form leader")
			timeoutTimer.Reset(timeout)
			continue
		}
	}
	return role{}
}

func Start() {
	r := newRole()
	log.Debugf("current role is %v", r)
	rand.Seed(42)
	r.f()

}

func sendHeartBeatsForever() {
	for {
		select {
		case <-time.After(time.Second * time.Duration(HeartTimerBase)):
			log.Debugf("send heart beat")
			for _, server := range config.Config.Servers {
				go leaderHeartBeat(server)
			}
		}
	}
}

func candidateStage() int {
	gs.ReElect()
	gs.AddTerm()

	wg := sync.WaitGroup{}
	wg.Add(len(config.Config.Servers))
	log.Debugf("current state is: %+v", gs)
	ch := make(chan uint32, len(config.Config.Servers))
	for _, server := range config.Config.Servers {
		go func(s string) {
			defer wg.Done()
			ch <- requestForVote(s)
		}(server)
	}
	wg.Wait()

	var sum uint32
	for i := 0; i < len(config.Config.Servers); i++ {
		sum += <-ch
	}
	if sum >= gs.NodeCount/2+1 {
		log.Debug("win this vote, become leader, count:", sum)
		return leader
	}
	return candidate
}

func leaderHeartBeat(server string) {
	log.Debug("append entry to" + server)
	args := &rpcs.Args{Term: gs.GetTerm(), LeaderId: state.MyID, Entries: nil}
	reply := new(rpcs.Results)

	callRpc(server, args, reply, "Rpc.AppendEntries")

}

func requestForVote(server string) uint32 {
	log.Debug("rpc for vote, server is:", server)

	args := &rpcs.RequestVoteArgs{
		Term:         gs.GetTerm(),
		CandidateId:  state.MyID,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}
	reply := new(rpcs.RequestVoteReply)
	callRpc(server, args, reply, "Rpc.RequestVote")

	log.Debugf("rpc reply is: %+v from: %v", reply, server)
	if reply.VoteGranted {
		return 1
	}
	return 0
}

func callRpc(server string, args interface{}, reply interface{}, s string) error {
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		log.Error("dialing:", err)
		return err
	}

	err = client.Call(s, args, reply)
	if err != nil {
		log.Error("call rpc error:", err)
	}

	return nil
}
