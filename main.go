package main

import (
	"os"
	"os/signal"
	_ "raft/config"
	"raft/election"
	_ "raft/rpcs"
)

func main() {
	election.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
