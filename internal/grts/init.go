package grts

import (
	"log/slog"

	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
	"github.com/Morgahl/gotp/supervisor/static"
)

var _ supervisor.Supervisable = &Init{}

// var init_OPTIONS = supervisor.Options{
// 	AutoShutdown: supervisor.NEVER,
// 	Strategy:     supervisor.ONE_FOR_ONE,
// }

type Init struct{}

func (i *Init) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		ID:          "grts.Init",
		Restart:     supervisor.PERMANENT,
		Shutdown:    0,
		Type:        supervisor.SUPERVISOR,
		Significant: true,
		Start: func(opts ...process.SpawnOpt) (process.Ref, error) {
			opts = append([]process.SpawnOpt{process.Named("grts.Init")}, opts...)
			return static.Start(i, nil, opts...)
		},
	}
}

func (i *Init) Init(pctx process.Context, opts process.Options) (supervisor.Options, []supervisor.Supervisable, error) {
	slog.DebugContext(pctx.Context(), "grts.Init.Init 