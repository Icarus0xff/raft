package main

import (
	"os"
	"os/signal"
	_ "raft/config"
	"raft/rpcs"
	_ "raft/rpcs"
	"raft/stateMachine"
)

func main() {

	sm := stateMachine.NewRoleStateMachine(rpcs.RealCall{})
	sm.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
