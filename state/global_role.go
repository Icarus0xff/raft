package state

const (
	Follower Role = iota
	Candidate
	Leader
)

type Role int32

func (r *Role) AmILeader() bool {
	return *r == Leader
}

var gr *Role

func GetGlobalRolePtr() *Role {
	return gr
}

func init() {
	gr = new(Role)
}
