package game

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	gotp "github.com/Morgahl/gotp/old"
	"github.com/Morgahl/gotp/old/server"
)

const (
	MIN_DURATION = 500 * time.Microsecond
	MID_DURATION = 5 * time.Millisecond
	MAX_DURATION = 500 * time.Millisecond
)

var _ gotp.Supervisable = &Crew{}
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

func (f *Crew) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		ID:       f.id,
		Restart:  gotp.TRANSIENT,
		Shutdown: gotp.DEFAULT_SHUTDOWN,
		Type:     gotp.WORKER,
	}
}

func (f *Crew) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return f.server.Start(opts...)
}

func (f *Crew) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (s gotp.Supervised, err error) {
	return f.server.StartLink(link, opts...)
}

func (f *Crew) Init(opts gotp.Options) (c server.Continue[any], err error) {
	start := time.Now()
	slog.Debug("Crew.Init", "id", f.id, "pid", f.server.ID(), "took", time.Since(start), "opts", opts)
	f.wi = workItem{
		id:  f.id,
		rem: rand.Intn(100) + 101,
	}
	f.server.Send(server.CastMsg(f.wi))
	return server.NoCont[any](), nil
}

func (f *Crew) HandleCast(work workItem) (server.Continue[any], error) {
	if work != f.wi {
		return server.NoCont[any](), fmt.Errorf("unexpected work item: %v", work)
	}
	f.wi.rem--
	load := assessWork()
	time.Sleep(load)
	f.wi.taken += load
	if f.wi.rem >= 1 {
		f.server.Send(server.CastMsg(f.wi))
	} else {
		f.server.Stop(nil)
		return server.NoCont[any](), nil
	}
	return server.NoCont[any](), nil
}

func (f *Crew) Terminate(reason error) error {
	if f.wi.rem > 0 {
		slog.Error("Crew.Terminate", "id", f.id, "pid", f.server.ID(), "work", f.wi, "reason", reason)
	} else {
		slog.Info("Crew.Terminate", "id", f.id, "pid", f.server.ID(), "work", f.wi, "reason", reason)
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
	case n <= 0.6:
		return time.Duration(rand.Int63n(int64(MIN_DURATION)))
	case n <= 0.8:
		return time.Duration(rand.Int63n(int64(MID_DURATION-MIN_DURATION))) + MIN_DURATION
	case n <= 0.95:
		return time.Duration(rand.Int63n(int64(MAX_DURATION-MID_DURATION))) + MID_DURATION
	default:
		return time.Duration(rand.Int63n(int64(MAX_DURATION))) + MID_DURATION
	}
}
