package rpcs

import (
	"github.com/ngaut/log"
	"raft/state"
)

type Args struct {
	Term         uint32
	LeaderId     uint32
	prevLogIndex int
	prevLogTerm  int
	Entries      []int
	leaderCommit int
}

type Results struct {
	Term     int
	IsSucess bool
}

type Rpc int

func (r *Rpc) AppendEntries(args *Args, reply *Results) error {
	log.Debug("AppendEntries")
	*reply = Results{1, true}
	log.Debug("AppendEntries end")
	return nil
}

type RequestVoteArgs struct {
	Term         uint32
	CandidateId  uint32
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        uint32
	VoteGranted bool
}

var gs = state.GetGlobalState()

func (r *Rpc) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	log.Debugf("args is: %+v", *args)
	*reply = RequestVoteReply{Term: gs.GetTerm(), VoteGranted: false}
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
