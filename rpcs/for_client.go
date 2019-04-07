package rpcs

import (
	"github.com/ngaut/log"
	"raft/config"
	. "raft/logStruct"
	. "raft/state"
)

type CommandReply struct {
	IsSucceed bool
}

func (r *Rpc) SetGetDelKV(args *OperationLogEntry, reply *CommandReply) error {
	log.Debugf("SetGetDelKV args is %+v", *args)
	rr := GetGlobalRolePtr()
	s := GetGlobalState()

	args.Term = s.CurrentTerm()
	if rr.AmILeader() {
		s.Log = append(s.Log, *args)
	} else {
		log.Debugf("i'm not a leader, can't do it")
		return nil
	}

	for _, s := range config.Config.Servers {
		log.Debug("server", s)

		args := &AppendEntriesArgs{
			Term:         args.Term,
			LeaderId:     int32(MyID),
			prevLogIndex: len(GetGlobalState().Log) - 1,
			prevLogTerm:  1,
			Entries:      []OperationLogEntry{*args},
			leaderCommit: 1,
		}
		reply := new(AppendEntriesReply)

		client := RealCall{}

		err := client.Call("Rpc.AppendEntries", s, args, reply)
		if err != nil {
			log.Error("call Rpc.AppendEntries error:", err)
			return err
		}
		log.Debugf("call successul", s)
	}

	return nil
}
