package election

import (
	"github.com/ngaut/log"
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

type role int

var r role
var gs = state.GetGlobalState()

const electionTimeout = time.Second * 20

func Start() {
	log.Debugf("current role: %v, state: %+v", r, gs)

	select {
	case <-time.After(electionTimeout):
		log.Debug("follower timeout, prepare to vote.")
	}
	candidateStage()
	if r == leader {
		for _, server := range config.Config.Servers {
			go leaderHeartBeat(server)
		}
	}
}

func candidateStage() {
	gs.ReElect()
	r = candidate
	gs.AddTerm()

	wg := sync.WaitGroup{}
	wg.Add(len(config.Config.Servers))
	log.Debugf("current state is: %+v", gs)
	for _, server := range config.Config.Servers {
		go func(s string) {
			defer wg.Done()
			requestForVote(s)
		}(server)
	}
	wg.Wait()
	log.Debug("end candidateStage")
}

func leaderHeartBeat(server string) {
	log.Debug("append entry")
	args := &rpcs.Args{Term: gs.GetTerm(), LeaderId: state.MyID, Entries: nil}
	reply := new(rpcs.Results)

	callRpc(server, args, reply, "Rpc.AppendEntries")

}

func requestForVote(server string) {
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
		c := gs.GetVoteFromCandidate()
		if c >= gs.NodeCount/2+1 {
			r = leader
			log.Debug("win this vote, become leader, count:", c)
		}
	}
}

func callRpc(server string, args interface{}, reply interface{}, s string) {
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call(s, args, reply)
	if err != nil {
		log.Error("call Rpc.RequestVote error:", err)
	}
}
