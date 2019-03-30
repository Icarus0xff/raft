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
	err := rpc.Register(r)
	if err != nil {
		log.Fatal("register rpc server error", err)
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", config.Config.GetBindAddr())
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go func() {
		err := http.Serve(l, nil)
		if err != nil {
			log.Fatal("serve rpc error", err)
		}
	}()
}
