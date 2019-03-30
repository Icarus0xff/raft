package rpcs

import (
	"github.com/ngaut/log"
	"raft/state"
)

type AppendEntriesArgs struct {
	Term         uint32
	LeaderId     uint32
	prevLogIndex int
	prevLogTerm  int
	Entries      []int
	leaderCommit int
}

type AppendEntriesReply struct {
	Term      uint32
	IsSucceed bool
}

type Rpc int

func (r *Rpc) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	log.Debug("AppendEntries")
	reply = new(AppendEntriesReply)
	if args.Term < state.GetGlobalState().GetTerm() {
		*reply = AppendEntriesReply{state.GetGlobalState().GetTerm(), false}
		return nil

	}
	*reply = AppendEntriesReply{1, true}
	state.GetGlobalState().LeaderHeartBeat <- struct{}{}
	log.Debug("AppendEntries end")
	return nil
}

type ReqVoteArgs struct {
	Term         uint32
	CandidateId  uint32
	LastLogIndex int
	LastLogTerm  int
}

type ReqVoteReply struct {
	Term        uint32
	VoteGranted bool
}

var gs = state.GetGlobalState()

func (r *Rpc) RequestVote(args *ReqVoteArgs, reply *ReqVoteReply) error {
	log.Debugf("args is: %+v", *args)
	*reply = ReqVoteReply{Term: gs.GetTerm(), VoteGranted: false}
	if args.Term < gs.GetTerm() {
		log.Debug("term is less than current term")
		return nil
	}

	if gs.IsVoted() {
		reply.VoteGranted = false
		log.Debug("has voted for this term")
	} else {
		reply.VoteGranted = true

		gs.VoteCandidate(args.CandidateId)
		log.Debug("vote candidate:", args.CandidateId)
	}
	log.Debugf("reply is: %+v", *reply)
	return nil
}
