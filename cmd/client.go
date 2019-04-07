package main

import (
	"github.com/ngaut/log"
	. "raft/logStruct"
	"raft/rpcs"
)

func main() {
	log.Debug("debug")

	args := OperationLogEntry{
		Operation: PUT,
		Key:       "fuck",
		Value:     "shit",
	}
	reply := new(rpcs.CommandReply)

	client := rpcs.RealCall{}

	err := client.Call("Rpc.SetGetDelKV", "localhost:2002", args, reply)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("reply is: %+v", reply)
	return
}
