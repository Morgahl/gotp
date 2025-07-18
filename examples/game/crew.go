package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

const (
	MIN_DURATION      = 2 * time.Second
	MID_LOW_DURATION  = 3 * time.Second
	MID_HIGH_DURATION = 5 * time.Second
	MAX_DURATION      = 8 * time.Second

	atom_GET_WORK    gotp.Atom = "get_work"
	atom_SUBMIT_WORK gotp.Atom = "submit_work"
)

var _ server.Supervisable = &Crew{}
var _ server.Serverable[gotp.Options, any, any, *workItem, any] = &Crew{}

type Crew struct {
	id    gotp.Atom
	agent gotp.Atom

	// Embed the server.DefaultHandlers to provide default implementations
	// for the server.Serverable interface methods.
	server.OptionalCallbacks[any, any]
	server *server.Server[gotp.Options, any, any, *workItem, any]
}

func NewCrew(id gotp.Atom, agent gotp.Atom) *Crew {
	c := Crew{id: id, agent: agent}
	c.server = server.New(&c, nil)
	return &c
}

func (f *Crew) Context() context.Context {
	return f.server.Process().Context()
}

func (f *Crew) ChildSpec() server.ChildSpec {
	return server.ChildSpec{
		ID:        f.id,
		Restart:   server.TRANSIENT,
		Shutdown:  gotp.DEFAULT_SHUTDOWN,
		Type:      server.WORKER,
		SpawnOpts: []process.SpawnOpt{process.Named(f.id)},
	}
}

func (f *Crew) Start(opts ...process.SpawnOpt) (process.Started, error) {
	return f.server.Start(opts...)
}

func (f *Crew) StartLink(linked *process.Process, opts ...process.SpawnOpt) (s server.Supervised, err error) {
	return f.server.StartLink(linked, opts...)
}

func (f *Crew) Init(opts gotp.Options) (c server.Continue[any], err error) {
	// slog.InfoContext(f.Context(), "Crew.Init", slog.Any("agent", f.agent))
	f.server.Process().UpdateFlags(func(flags process.ProcessFlags) process.ProcessFlags {
		flags |= process.TRAP_EXIT_FLAG
		return flags
	})
	return server.Cont[any](atom_GET_WORK), nil
}

func (f *Crew) HandleContinue(msg any) (server.Continue[any], error) {
	switch m := msg.(type) {
	case gotp.Atom:
		switch m {
		case atom_GET_WORK:
			if w, ok := GetWork(f.agent, f.server.PID()); ok {
				// slog.DebugContext(f.Context(), "Crew.HandleContinue", slog.Any("work", w))
				f.server.Send(server.CastMsg(w))
				return server.NoCont[any](), nil
			}
			return server.Stop[any](process.NORMAL), nil
		}
	}
	return server.NoCont[any](), nil
}

func (f *Crew) HandleCast(work *workItem) (server.Continue[any], error) {
	if work.need != work.done {
		work.done++
		load := assessWork()
		work.taken += load
		f.server.SendAfter(server.CastMsg(work), load)
		return server.NoCont[any](), nil
	}

	SubmitProcessedWork(f.agent, work)
	return server.Cont[any](atom_GET_WORK), nil
}

func (f *Crew) HandleInfo(msg process.Message) (server.Continue[any], error) {
	switch m := msg.(type) {
	case process.ExitMsg:
		if m.PID == f.server.PID() {
			return server.Stop[any](m.Reason), nil
		}
	default:
		return server.NoCont[any](), fmt.Errorf("unexpected message: %T", m)
	}
	return server.NoCont[any](), nil
}

func (f *Crew) Terminate(reason error) error {
	switch {
	case errors.Is(reason, process.NORMAL) || errors.Is(reason, ctx.Shutdown{}):
		slog.InfoContext(f.Context(), "Crew.Terminate", slog.Any("reason", reason))
	default:
		slog.ErrorContext(f.Context(), "Crew.Terminate", slog.Any("reason", reason))
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
