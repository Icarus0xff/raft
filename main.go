package main

import (
	"os"
	"os/signal"
	_ "raft/config"
	_ "raft/rpcs"
	"raft/stateMachine"
)

func main() {

	sm := stateMachine.NewRoleStateMachine()
	sm.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
