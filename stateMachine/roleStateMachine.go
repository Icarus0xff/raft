package stateMachine

import (
	"github.com/ngaut/log"
	"math/rand"
	"raft/config"
	"raft/rpcs"
	"raft/state"
	"raft/utils"
	"sync"
	"time"
)

const (
	follower = iota
	candidate
	leader
)

// heartTimerBase must less than timeoutBase
const timeoutBase = 3
const heartTimerBase = 2

var gs = state.GetGlobalState()

type RoleStateMachine struct {
	role int
}

func NewRoleStateMachine() *RoleStateMachine {
	return &RoleStateMachine{follower}
}

func (r *RoleStateMachine) Run() {
	for {
		switch r.role {
		case follower:
			r.followerStage().Run()
		case candidate:
			r.candidateStage().Run()
		case leader:
			r.leaderStage().Run()
		}
	}

}

func (r *RoleStateMachine) followerStage() *RoleStateMachine {
	log.Debug("follower stage")

	delta, timeoutTimer := r.randomTimer()

	log.Debugf("random delta is %+v", delta)

	<-timeoutTimer.C
	return &RoleStateMachine{candidate}
}

func (r *RoleStateMachine) candidateStage() *RoleStateMachine {
	gs.ElectInit()
	gs.AddTerm()
	if isReqVotesSucceed() {
		return &RoleStateMachine{leader}
	}

	_, timeoutTimer := r.randomTimer()
	<-timeoutTimer.C
	return &RoleStateMachine{candidate}
}

func (r *RoleStateMachine) leaderStage() *RoleStateMachine {
	select {
	case <-time.After(time.Second * time.Duration(heartTimerBase)):
		log.Debugf("send heart beat")
		for _, server := range config.Config.Servers {
			go sendHeartBeat(server)
		}
	}
	return &RoleStateMachine{leader}
}

func (r *RoleStateMachine) randomTimer() (time.Duration, *time.Timer) {
	ri := rand.Intn(10) + timeoutBase
	timeout := time.Second * time.Duration(ri)
	timeoutTimer := time.NewTimer(timeout)
	return timeout, timeoutTimer
}

func isReqVotesSucceed() bool {
	log.Debugf("current state is: %+v", gs)

	wg := sync.WaitGroup{}
	sc := len(config.Config.Servers)
	wg.Add(sc)

	vsc := make(chan uint32, sc)
	for _, server := range config.Config.Servers {
		go func(s string) {
			defer wg.Done()
			vsc <- requestForVote(s)
		}(server)
	}
	wg.Wait()

	var sum uint32
	for i := 0; i < sc; i++ {
		sum += <-vsc
	}
	if sum >= gs.NodeCount/2+1 {
		log.Debug("win this vote, become leader, count:", sum)
		return true
	}
	return false
}

func sendHeartBeat(server string) {
	log.Debug("append entry to" + server)
	args := &rpcs.Args{Term: gs.GetTerm(), LeaderId: state.MyID, Entries: nil}
	reply := new(rpcs.Results)

	err := utils.CallRpc(server, args, reply, "Rpc.AppendEntries")
	if err != nil {
		log.Error("send heart beat error", err)
	}
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
	err := utils.CallRpc(server, args, reply, "Rpc.RequestVote")
	if err != nil {
		log.Error("send vote request error", err)
	}

	log.Debugf("rpc reply is: %+v from: %v", reply, server)
	if reply.VoteGranted {
		return 1
	}
	return 0
}
