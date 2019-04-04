package state

import (
	"github.com/ngaut/log"
	"hash/fnv"
	"raft/config"
	"sync/atomic"
)

var MyID uint32

var gs *State

func GetGlobalState() *State {
	return gs
}

func init() {
	gs = newState()
	hash := fnv.New32()
	log.Debug("if the BindAddr generated:", config.Config.GetBindAddr())
	_, err := hash.Write([]byte(config.Config.GetHostAddr()))
	if err != nil {
		log.Fatal("gen hash error", err)
	}
	MyID = hash.Sum32()
	log.Debug("my id is:", MyID)
}

type State struct {
	currentTerm uint32
	votedFor    int32
	Log         []string

	inner
}

type inner struct {
	myVotesCount    uint32
	NodeCount       uint32
	LeaderHeartBeat chan struct{}
	Role            Role
}

const notVoted = -1

func (s *State) ElectInit() {
	atomic.StoreInt32(&s.votedFor, notVoted)
	s.myVotesCount = 0
}

func (s *State) VoteForMyself() bool {
	return atomic.CompareAndSwapInt32(&s.votedFor, notVoted, int32(MyID))
}

func (s *State) VoteCandidate(id uint32) {
	atomic.StoreInt32(&s.votedFor, int32(id))
}

func (s *State) IsVoted() bool {
	return atomic.LoadInt32(&s.votedFor) != notVoted
}

func (s *State) AddTerm() uint32 {
	return atomic.AddUint32(&s.currentTerm, 1)
}

func (s *State) GetTerm() uint32 {
	return atomic.LoadUint32(&s.currentTerm)
}

func (s *State) GetVoteFromCandidate() uint32 {
	return atomic.AddUint32(&s.myVotesCount, 1)
}

func newState() *State {
	c := uint32(len(config.Config.Servers) + 1)
	return &State{0, notVoted, []string{},
		inner{0, c,
			make(chan struct{}), Follower},
	}
}
