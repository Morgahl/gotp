package static

import (
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

func Start[I any](sup supervisor.Supervisor[I], initArg I, opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.Start(&server[I]{sup: sup}, initArg, opts...)
}

func StartLink[I any](sup supervisor.Supervisor[I], initArg I, linked process.Ref, opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.StartLink(&server[I]{sup: sup}, initArg, linked, opts...)
}
