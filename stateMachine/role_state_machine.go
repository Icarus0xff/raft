package stateMachine

import (
	"github.com/ngaut/log"
	"math/rand"
	"raft/config"
	"raft/rpcs"
	. "raft/state"
	"sync"
	"time"
)

// heartTimerBase must less than timeoutBase
const timeoutBase = 3
const heartTimerBase = 2

type RoleStateMachine struct {
	Role        *Role
	client      rpcs.Callee
	globalState *State

	candidateTimeout *time.Timer
}

func NewRoleStateMachineDefault() *RoleStateMachine {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	timer := time.NewTimer(timeout)

	return &RoleStateMachine{setFollower(), rpcs.RealCall{},
		GetGlobalState(), timer}
}

func setFollower() *Role {
	r := GetGlobalRolePtr()
	*r = Follower
	return r
}

func NewRoleStateMachine(client rpcs.Callee) *RoleStateMachine {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	timer := time.NewTimer(timeout)

	return &RoleStateMachine{setFollower(), client,
		GetGlobalState(), timer}
}

func (r *RoleStateMachine) StartServer(addr string) {
	a := ""
	if addr == "" {
		a = config.Config.GetBindAddr()
	} else {
		a = addr
	}
	rpcs.RunServer(a)
	r.run()
}
func (r *RoleStateMachine) run() {

	for {
		switch *r.Role {
		case Follower:
			r.followerStage().run()
		case Candidate:
			r.candidateStage().run()
		case Leader:
			r.leaderStage().run()
		}
	}

}

func (r *RoleStateMachine) resetTimer() time.Duration {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	r.candidateTimeout.Reset(timeout)
	return timeout
}

// 如果收到心跳包则重置当前的选举timer
func (r *RoleStateMachine) followerStage() *RoleStateMachine {
	log.Debug("follower stage")

	return r.candidateFollowerMachine()
}

func (r *RoleStateMachine) candidateStage() *RoleStateMachine {
	r.globalState.ElectInit()
	r.globalState.AddTerm()
	log.Debugf("added term, current term is", r.globalState.CurrentTerm())

	GetGlobalState().VoteForMyself()
	if r.isReqVotesSucceed() {
		*r.Role = Leader
		return r
	}

	return r.candidateFollowerMachine()
}

func (r *RoleStateMachine) candidateFollowerMachine() *RoleStateMachine {
	timeout := r.resetTimer()
	select {
	case <-r.globalState.LeaderHeartBeat:
		r.candidateTimeout.Reset(timeout)
		log.Debugf("recv heartbeat")
		*r.Role = Follower
		return r
	case <-r.candidateTimeout.C:
		log.Info("candidate timeout")
		*r.Role = Candidate
		return r
	}
}

func (r *RoleStateMachine) leaderStage() *RoleStateMachine {
	select {
	case <-time.After(time.Second * time.Duration(heartTimerBase)):
		log.Debugf("send heart beat")
		for _, server := range config.Config.Servers {
			go r.sendHeartBeat(server)
		}
	}
	*r.Role = Leader
	return r
}

func (r *RoleStateMachine) randomTimer() (time.Duration, *time.Timer) {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	timeoutTimer := time.NewTimer(timeout)
	return timeout, timeoutTimer
}

func (r *RoleStateMachine) isReqVotesSucceed() bool {
	log.Debugf("current state is: %+v", r.globalState)

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

	var sum uint32 = 1
	for i := 0; i < sc; i++ {
		sum += <-vsc
	}
	if sum >= r.globalState.NodeCount/2+1 {
		log.Debug("win this term, become leader, count:", sum)
		return true
	}
	return false
}

func (r *RoleStateMachine) sendHeartBeat(server string) {
	log.Debug("append entry to" + server)
	args := &rpcs.AppendEntriesArgs{Term: r.globalState.CurrentTerm(),
		LeaderId: int32(MyID), Entries: nil}
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
		Term:         r.globalState.CurrentTerm(),
		CandidateId:  MyID,
		LastLogIndex: 1,
		LastLogTerm:  1,
	}
	reply := new(rpcs.ReqVoteReply)
	err := r.client.Call("Rpc.RequestVote", server, args, reply)
	if err != nil {
		log.Error("send vote request error", err)
		return 0
	}

	log.Debugf("rpc reply is: %+v from: %v", reply, server)
	if reply.VoteGranted {
		return 1
	}
	return 0
}
