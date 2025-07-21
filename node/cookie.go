package node

// Cookie is a type used to represent a preshared secret for authentication. This both acts to secure the connection
// with an environmentally injected secret and limits accidental cross-talk between nodes should network configuration
// accidentally allow communication between nodes that should not be able to communicate.
//
// OF NOTE: A node may have a local `Cookie` this is used for all incoming connections, and by default when dialing
// another node the local `Cookie` will be used. Howver it should be noted that `Node` and `Cookies` are paired, so when
// dilaing another node you may specify a different `Cookie` to use for that connection.
type Cookie string

func (c Cookie) String() string {
	return "****************"
}

func (c Cookie) GoString() string {
	return c.String()
}
