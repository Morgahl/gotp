package foo

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
)

const (
	MIN_DURATION = 10 * time.Millisecond
	MID_DURATION = 50 * time.Millisecond
	MAX_DURATION = 200 * time.Millisecond
)

var _ gotp.Supervisable = &FooServer{}
var _ gen_server.Serverable[gotp.Options, any, any, any, any] = &FooServer{}

type FooServer struct {
	id   gotp.Atom
	work time.Duration

	// Embed the gen_server.DefaultHandlers to provide default implementations
	// for the gen_server.Serverable interface methods.
	gen_server.OptionalCallbacks[any, any]
}

func NewFooServer(id gotp.Atom) *FooServer {
	return &FooServer{
		id:   id,
		work: assessWork(),
	}
}

func (f *FooServer) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		ID:       f.id,
		Restart:  gotp.PERMANENT,
		Shutdown: 30 * time.Second,
		Type:     gotp.WORKER,
	}
}

func (f *FooServer) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return gen_server.New(f, nil).Start(opts...)
}

func (f *FooServer) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return gen_server.New(f, nil).StartLink(link, opts...)
}

func (f *FooServer) Init(opts gotp.Options) (gen_server.Continue[any], error) {
	start := time.Now()
	simulateWork(f.work)
	slog.Info("FooServer.Init", "id", f.id, "opts", opts, "took", time.Since(start))
	return gen_server.NoCont[any](), nil
}

func (f *FooServer) Terminate(reason error) error {
	start := time.Now()
	simulateWork(f.work)
	slog.Info("FooServer.Terminate", "id", f.id, "reason", reason, "took", time.Since(start))
	return reason
}

func assessWork() time.Duration {
	switch n := rand.Float64(); {
	case n <= 0.3:
		return time.Duration(rand.Int63n(int64(MIN_DURATION)))
	case n <= 0.7:
		return time.Duration(rand.Int63n(int64(MID_DURATION-MIN_DURATION))) + MIN_DURATION
	case n <= 0.9:
		return time.Duration(rand.Int63n(int64(MAX_DURATION-MID_DURATION))) + MID_DURATION
	default:
		return time.Duration(rand.Int63n(int64(MAX_DURATION))) + MID_DURATION
	}
}

func simulateWork(work time.Duration) {
	// halfWork := work / 2
	// time.Sleep(time.Duration(rand.Int63n(int64(work-halfWork))) + halfWork)
}
