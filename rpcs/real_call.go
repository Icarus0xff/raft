package rpcs

import (
	"github.com/ngaut/log"
	"net/rpc"
)

type RealCall struct {
}

func (c RealCall) Call(name, server string, args, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", server)
	if err != nil {
		log.Error("dialing:", err)
		return err
	}

	err = client.Call(name, args, reply)
	if err != nil {
		log.Error("call rpc error:", err)
	}

	return nil
}
