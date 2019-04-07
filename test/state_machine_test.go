package test

import (
	"github.com/ngaut/log"
	"github.com/stretchr/testify/assert"
	"raft/config"
	"raft/logStruct"
	"raft/rpcs"
	"raft/state"
	"raft/stateMachine"
	"testing"
)

func TestStateMachine(t *testing.T) {

	fc := rpcs.NewFakeCall()
	sm := stateMachine.NewRoleStateMachine(fc)

	go sm.StartServer("")

	log.Debug("waiting for chan")
	<-fc.Ch
	assert.Equal(t, *sm.Role, state.Candidate)
	return
}

func TestServiceCall(t *testing.T) {

	log.Debug("test service")
	sm := stateMachine.NewRoleStateMachineDefault()

	go sm.StartServer("")

	args := &logStruct.OperationLogEntry{
		Operation: logStruct.PUT,
		Key:       "fuck",
		Value:     "shit",
	}
	reply := new(rpcs.CommandReply)

	client := rpcs.RealCall{}

	client.Call("Rpc.SetGetDelKV", config.Config.GetBindAddr(), args, reply)

	return
}
