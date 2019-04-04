package stateMachine

import (
	"github.com/ngaut/log"
	"math/rand"
	"raft/config"
	"raft/rpcs"
	"raft/state"
	"sync"
	"time"
)

// heartTimerBase must less than timeoutBase
const timeoutBase = 3
const heartTimerBase = 2

type RoleStateMachine struct {
	Role   *state.Role
	client rpcs.Callee
	state  *state.State
}

func NewRoleStateMachineDefault() *RoleStateMachine {
	r := state.GetGlobalRolePtr()
	*r = state.Follower
	return &RoleStateMachine{r, rpcs.RealCall{}, state.GetGlobalState()}
}

func NewRoleStateMachine(client rpcs.Callee) *RoleStateMachine {
	r := state.GetGlobalRolePtr()
	*r = state.Follower
	return &RoleStateMachine{r, client, state.GetGlobalState()}
}

func (r *RoleStateMachine) Run() {
	for {
		switch *r.Role {
		case state.Follower:
			r.followerStage().Run()
		case state.Candidate:
			r.candidateStage().Run()
		case state.Leader:
			r.leaderStage().Run()
		}
	}

}

func (r *RoleStateMachine) followerStage() *RoleStateMachine {
	log.Debug("follower stage")

	delta, timeoutTimer := r.randomTimer()

	log.Debugf("random delta is %+v", delta)

	<-timeoutTimer.C

	*r.Role = state.Candidate
	return r
}

func (r *RoleStateMachine) candidateStage() *RoleStateMachine {
	r.state.ElectInit()
	r.state.AddTerm()
	if r.isReqVotesSucceed() {
		*r.Role = state.Leader
		return r
	}

	_, timeoutTimer := r.randomTimer()
	<-timeoutTimer.C
	*r.Role = state.Candidate
	if state.GetGlobalState().VoteForMyself() {
		state.GetGlobalState().GetVoteFromCandidate()
	}
	return r
}

func (r *RoleStateMachine) leaderStage() *RoleStateMachine {
	select {
	case <-time.After(time.Second * time.Duration(heartTimerBase)):
		log.Debugf("send heart beat")
		for _, server := range config.Config.Servers {
			go r.sendHeartBeat(server)
		}
	}
	*r.Role = state.Leader
	return r
}

func (r *RoleStateMachine) randomTimer() (time.Duration, *time.Timer) {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	timeoutTimer := time.NewTimer(timeout)
	return timeout, timeoutTimer
}

func (r *RoleStateMachine) isReqVotesSucceed() bool {
	log.Debugf("current state is: %+v", r.state)

	wg := sync.WaitGroup{}
	sc := len(config.Config.Servers)
	wg.Add(sc)

	vsc := make(chan uint32, sc)
	for _, server := range config.Config.Servers {
		go func(s string) {
			defer wg.Done()
			vsc <- r.requestForVote(s)
		}(server)
	}
	wg.Wait()

	var sum uint32
	for i := 0; i < sc; i++ {
		sum += <-vsc
	}
	if sum >= r.state.NodeCount/2+1 {
		log.Debug("win this term, become leader, count:", sum)
		return true
	}
	return false
}

func (r *RoleStateMachine) sendHeartBeat(server string) {
	log.Debug("append entry to" + server)
	args := &rpcs.AppendEntriesArgs{Term: r.state.GetTerm(),
		LeaderId: int32(state.MyID), Entries: nil}
	reply := new(rpcs.AppendEntriesReply)

	rpc := rpcs.RealCall{}
	err := rpc.Call("Rpc.AppendEntries", server, args, reply)
	if err != nil {
		log.Error("send heart beat error", err)
	}
}

func (r *RoleStateMachine) requestForVote(server string) uint32 {
	log.Debug("rpc for vote, server is:", server)

	args := &rpcs.ReqVoteArgs{
		Term:         r.state.GetTerm(),
		CandidateId:  state.MyID,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}
	reply := new(rpcs.ReqVoteReply)
	err := r.client.Call("Rpc.RequestVote", server, args, reply)
	if err != nil {
		log.Error("send vote request error", err)
	}

	log.Debugf("rpc reply is: %+v from: %v", reply, server)
	if reply.VoteGranted {
		return 1
	}
	return 0
}
