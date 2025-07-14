package game

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

const (
	// MIN_DURATION = 500 * time.Microsecond
	// MID_DURATION = 5 * time.Millisecond
	// MAX_DURATION = 500 * time.Millisecond

	MIN_DURATION = 2 * time.Second
	MID_DURATION = 3 * time.Second
	MAX_DURATION = 5 * time.Second
)

var _ server.Supervisable = &Crew{}
var _ server.Serverable[gotp.Options, any, any, workItem, any] = &Crew{}

type Crew struct {
	id gotp.Atom
	wi workItem

	// Embed the server.DefaultHandlers to provide default implementations
	// for the server.Serverable interface methods.
	server.OptionalCallbacks[any, any]
	server *server.Server[gotp.Options, any, any, workItem, any]
}

func NewCrew(id gotp.Atom) *Crew {
	c := Crew{id: id}
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
	f.server.Process().UpdateFlags(func(flags process.ProcessFlags) process.ProcessFlags {
		flags |= process.TRAP_EXIT_FLAG
		return flags
	})
	start := time.Now()
	slog.DebugContext(f.Context(), "Crew.Init", "took", time.Since(start), "opts", opts)
	f.wi = workItem{
		id: f.id,
		// rem: rand.Intn(15) + 16,
		rem: 1,
	}
	f.server.Send(server.CastMsg(f.wi))
	return server.NoCont[any](), nil
}

func (f *Crew) HandleCast(work workItem) (server.Continue[any], error) {
	if work != f.wi {
		debug.Assert(f.wi.rem == 0, "Crew.HandleCast unexpected work item expected %v, got %v", f.wi, work)
		slog.InfoContext(f.Context(), "Crew.HandleCast new work item", "work", work)
		f.wi = work
	}
	if f.wi.rem >= 1 {
		f.wi.rem--
		slog.DebugContext(f.Context(), "Crew.HandleCast working", "work", f.wi)
		load := assessWork()
		f.wi.taken += load
		f.server.SendAfter(server.CastMsg(f.wi), load)
	} else {
		return server.Stop[any](process.NORMAL), nil
	}
	return server.NoCont[any](), nil
}

func (f *Crew) HandleInfo(msg process.Message) (server.Continue[any], error) {
	slog.DebugContext(f.Context(), "Crew.HandleInfo", "msg", msg)
	switch m := msg.(type) {
	case process.ExitMsg:
		if m.PID == f.server.PID() {
			slog.InfoContext(f.Context(), "Crew.HandleInfo", "exit", m)
			return server.Stop[any](m.Reason), nil
		}
	default:
		slog.WarnContext(f.Context(), "Crew.HandleInfo", "unexpected", m)
		return server.NoCont[any](), fmt.Errorf("unexpected message: %T", m)
	}
	return server.NoCont[any](), nil
}

func (f *Crew) Terminate(reason error) error {
	if f.wi.rem > 0 {
		slog.ErrorContext(f.Context(), "Crew.Terminate", "work", f.wi, "reason", reason)
	} else {
		slog.InfoContext(f.Context(), "Crew.Terminate", "work", f.wi, "reason", reason)
	}
	return reason
}

type workItem struct {
	id    gotp.Atom
	taken time.Duration
	rem   int
}

func (w workItem) String() string {
	return fmt.Sprintf("workItem{id: %s, rem: %d}", w.id, w.rem)
}

func (w workItem) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("id", w.id),
		slog.Duration("taken", w.taken),
		slog.Int("rem", w.rem),
	)
}

func assessWork() time.Duration {
	switch n := rand.Float64(); {
	case n <= 0.33:
		return time.Duration(rand.Int63n(int64(MID_DURATION-MIN_DURATION))) + MIN_DURATION
	case n <= 0.67:
		return time.Duration(rand.Int63n(int64(MAX_DURATION-MID_DURATION))) + MID_DURATION
	default:
		return time.Duration(rand.Int63n(int64(MAX_DURATION))) + MID_DURATION
	}
}
