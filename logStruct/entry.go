package logStruct

type OPCode int

const (
	PUT OPCode = iota
)

type OperationLogEntry struct {
	Operation OPCode
	Key       string
	Value     string
	Term      uint32
}
