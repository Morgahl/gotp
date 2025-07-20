package game

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

const (
	MIN_DURATION      = 2 * time.Second
	MID_LOW_DURATION  = 3 * time.Second
	MID_HIGH_DURATION = 5 * time.Second
	MAX_DURATION      = 8 * time.Second

	atom_GET_WORK gotp.Atom = "get_work"
)

var _ supervisor.Supervisable = &Crew{}
var _ gen_server.GenServer[any, any, any, *workItem, any] = &Crew{}

func buildCrew(workAgent gotp.Atom, team gotp.Atom, crewList []gotp.Atom) []supervisor.Supervisable {
	var crew []supervisor.Supervisable
	for _, c := range crewList {
		crew = append(crew, NewCrew(team+"_"+c, workAgent))
	}
	return crew
}

type Crew struct {
	id    gotp.Atom
	agent gotp.Atom
	work  *workItem

	// Embed the gen_server.DefaultHandlers to provide default implementations
	// for the gen_server.Serverable interface methods.
	gen_server.OptionalCallbacks[any, any]
}

func NewCrew(id gotp.Atom, agent gotp.Atom) *Crew {
	c := Crew{id: id, agent: agent}
	return &c
}

func (f *Crew) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		ID:          f.id,
		Restart:     supervisor.TRANSIENT,
		Shutdown:    gotp.DEFAULT_SHUTDOWN,
		Type:        supervisor.WORKER,
		Significant: true,
		Start: func(opts ...process.SpawnOpt) (process.Ref, error) {
			return gen_server.Start(f, nil, append([]process.SpawnOpt{process.Named(f.id)}, opts...)...)
		},
	}
}

func (f *Crew) Init(pctx process.Context, _ any) (c gen_server.Continue[any], err error) {
	slog.DebugContext(pctx.Context(), "Crew.Init", slog.Any("agent", f.agent))
	pctx.TrapExit(true)
	return gen_server.Cont[any](atom_GET_WORK), nil
}

func (f *Crew) HandleContinue(pctx process.Context, msg any) (gen_server.Continue[any], error) {
	switch m := msg.(type) {
	case gotp.Atom:
		switch m {
		case atom_GET_WORK:
			if f.work == nil {
				if w, ok := ContextGetWork(pctx, f.agent, pctx.PID()); ok {
					f.work = w
					pctx.Send(gen_server.CastMsg(w))
					return gen_server.NoCont[any](), nil
				}

				slog.DebugContext(pctx.Context(), "Crew.HandleContinue - no work available")
				return gen_server.Stop[any](process.NORMAL), nil
			}
			work := f.work
			f.work = nil
			if work, ok := ContextSubmitProcessedWork(pctx, f.agent, work); ok {
				f.work = work
				pctx.Send(gen_server.CastMsg(work))
				return gen_server.NoCont[any](), nil
			}
			return gen_server.Stop[any](process.NORMAL), nil
		}
	}
	return gen_server.NoCont[any](), nil
}

func (f *Crew) HandleCast(pctx process.Context, work *workItem) (gen_server.Continue[any], error) {
	if f.work.id == work.id && work.need != work.done {
		work.done++
		load := assessWork()
		work.taken += load
		pctx.SendAfter(gen_server.CastMsg(work), load)
		return gen_server.NoCont[any](), nil
	}

	return gen_server.Cont[any](atom_GET_WORK), nil
}

func (f *Crew) HandleInfo(pctx process.Context, msg process.Message) (gen_server.Continue[any], error) {
	switch m := msg.(type) {
	case process.ExitMsg:
		if m.PID == pctx.PID() {
			return gen_server.Stop[any](m.Reason), nil
		}
	default:
		return gen_server.NoCont[any](), fmt.Errorf("unexpected message: %T", m)
	}
	return gen_server.NoCont[any](), nil
}

func (f *Crew) Terminate(pctx process.Context, reason error) error {
	switch {
	case errors.Is(reason, process.NORMAL) || errors.Is(reason, ctx.Shutdown{}):
		if f.work != nil {
			slog.ErrorContext(pctx.Context(), "Crew.Terminate", slog.Any("reason", reason), slog.Any("work_id", f.work))
		}
	default:
		slog.ErrorContext(pctx.Context(), "Crew.Terminate", slog.Any("reason", reason))
	}
	return reason
}

func assessWork() time.Duration {
	switch n := rand.Float64(); {
	case n <= 0.50:
		return time.Duration(rand.Int63n(int64(MID_LOW_DURATION-MIN_DURATION))) + MIN_DURATION
	case n <= 0.90:
		return time.Duration(rand.Int63n(int64(MID_HIGH_DURATION-MID_LOW_DURATION))) + MID_LOW_DURATION
	default:
		return time.Duration(rand.Int63n(int64(MAX_DURATION-MID_HIGH_DURATION))) + MID_HIGH_DURATION
	}
}
