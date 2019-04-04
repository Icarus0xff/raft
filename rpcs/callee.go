package rpcs

type Callee interface {
	Call(name, server string, args, reply interface{}) error
}
