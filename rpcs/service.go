package rpcs

import (
	"github.com/ngaut/log"
	. "raft/state"
)

type OPCode int

const (
	Put OPCode = iota
)

type CommandArgs struct {
	Operation OPCode
	Key       string
	Value     string
}

type CommandReply struct {
	IsSucceed bool
}

func (r *Rpc) Do(args *CommandArgs, reply *CommandReply) error {
	log.Debugf("args is %+v", *args)
	rr := GetGlobalRolePtr()
	if rr.AmILeader() {
		s := GetGlobalState()
		s.Log = append(s.Log, args.Key)
	}

	return nil
}
