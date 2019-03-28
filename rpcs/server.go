package rpcs

import (
	"github.com/ngaut/log"
	"net"
	"net/http"
	"net/rpc"
	"raft/config"
)

func init() {
	log.Debug("start rpc server")
	r := new(Rpc)
	rpc.Register(r)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", config.Config.GetBindAddr())
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go http.Serve(l, nil)
}
