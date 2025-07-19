package static

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

func Start(sup supervisor.Supervisor[gotp.Options], initArg gotp.Options, opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.Start(&server{sup: sup}, initArg, opts...)
}

func StartLink(sup supervisor.Supervisor[gotp.Options], initArg gotp.Options, linked process.Ref, opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.StartLink(&server{sup: sup}, initArg, linked, opts...)
}
