package gotp

type Node struct {
	name string
	host string
}

func (n Node) String() string {
	return n.name + "@" + n.host
}
