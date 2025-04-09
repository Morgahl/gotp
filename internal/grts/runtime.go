package grts

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
)

type Runtime struct {
	node gotp.Node
	apps map[string]application.Application
}
