package game

import (
	"expvar"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

type Metrics struct {
	Generated uint64
	Processed uint64
	Remaining uint64
}

func collectMetrics(agent *WorkAgent) func() any {
	return func() any {
		state, ok := GetState(agent.name, process.PIDZero())
		if !ok {
			return nil
		}
		return Metrics{
			Generated: state.generated,
			Processed: state.processed,
			Remaining: state.wanted - state.processed,
		}
	}
}

type WorkAgent struct {
	name   gotp.Atom
	wanted uint64
}

func NewWorkAgent(name gotp.Atom, wanted uint64) *WorkAgent {
	return &WorkAgent{
		name:   name,
		wanted: wanted,
	}
}

func (w *WorkAgent) ChildSpec() supervisor.ChildSpec {
	return agent.ChildSpec(
		func() *state {
			s := state{wanted: w.wanted, next: true}
			s.stepNextLog()
			expvar.Publish(string(w.name)+"_metrics", expvar.Func(collectMetrics(w)))
			return &s
		},
		process.Named(w.name),
		process.ChannelSize(max(runtime.GOMAXPROCS(0)<<3, 32)),
	)
}

func GetWork[S process.Sendable](agnt S, from process.PID) (*workItem, bool) {
	s, ok := agent.GetAndUpdate(agnt, from, func(state *state) state {
		state.generate()
		return *state
	}, 0) // indefinite timeout, this is an example that often runs at maximum capacity on systems to ensure "the system always moves forward"
	if !ok || !s.next {
		return nil, false
	}
	return s.workItem, s.next
}

func SubmitProcessedWork[S process.Sendable](agnt S, from process.PID, work *workItem) (*workItem, bool) {
	s, ok := agent.GetAndUpdate(agnt, from, func(state *state) state {
		if state.receivedProcessed(work) {
			slog.Warn("Submitted work", slog.Uint64("wanted", state.wanted), slog.Uint64("generated", state.generated), slog.Uint64("processed", state.processed), "from", from)
		}
		if state.processed == state.wanted {
			slog.Info("All work processed", slog.Uint64("wanted", state.wanted), slog.Uint64("generated", state.generated), slog.Uint64("processed", state.processed), "from", from)
			agent.Stop(agnt, process.NORMAL)
		}
		return *state
	}, 0) // indefinite timeout, this is an example that often runs at maximum capacity on systems to ensure "the system always moves forward"
	if !ok || !s.next {
		return nil, false
	}
	return s.workItem, s.next
}

func GetState[S process.Sendable](agnt S, from process.PID) (state, bool) {
	return agent.Get(agnt, from, func(state state) state { return state }, 0) // indefinite timeout, this is an example that often runs at maximum capacity on systems to ensure "the system always moves forward"
}

type state struct {
	wanted, generated, processed, nextLog uint64
	next                                  bool
	workItem                              *workItem
}

func (s *state) generate() {
	if !s.next {
		return
	}
	if s.generated >= s.wanted {
		s.workItem = nil
		s.next = false
		return
	}
	s.generated++
	s.next = true
	s.workItem = &workItem{
		id:   gotp.Atom(fmt.Sprintf("work-%d", s.generated)),
		need: uint64(rand.Intn(6) + 5),
	}
}

func (s *state) receivedProcessed(work *workItem) (log bool) {
	if work.need == work.done {
		s.processed++
		s.generate()
	}
	if s.processed >= s.nextLog {
		s.stepNextLog()
		return true
	}
	return false
}

// func (s *state) stepNextLog() {
// 	if s.wanted == 0 || s.processed >= s.wanted {
// 		s.nextLog = s.wanted + 1
// 		return
// 	}
// 	base := s.processed
// 	pos := float64(s.processed) / float64(s.wanted)
// 	if pos < 0 {
// 		pos = 0
// 	} else if pos > 1 {
// 		pos = 1
// 	}
// 	maxI := float64(s.wanted) * 0.01
// 	minI := 1.0
// 	easeIn := math.Pow(pos, 4.0)
// 	interval := maxI - (maxI-minI)*easeIn
// 	step := uint64(math.Max(1, math.Round(interval)))
// 	s.nextLog = base + step
// }

func (s *state) stepNextLog() {
	// we step by every 0.01% unless e are less then 10K work items in which case we step by every 100
	if s.wanted == 0 || s.processed >= s.wanted {
		s.nextLog = s.wanted + 1
		return
	}
	step := uint64(math.Max(1, math.Round(float64(s.wanted)*0.01)))
	if s.wanted <= 10000 {
		step = 100
	}
	s.nextLog = s.processed + step
}

type workItem struct {
	id    gotp.Atom
	taken time.Duration
	need  uint64
	done  uint64
}

func (w workItem) String() string {
	return fmt.Sprintf("workItem{id: %s, taken: %s, need: %d, done: %d, remaining: %s}",
		w.id, w.taken, w.need, w.done, w.remainingTime())
}

func (w workItem) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("id", w.id),
		slog.Duration("taken", w.taken),
		slog.Uint64("need", w.need),
		slog.Uint64("done", w.done),
		slog.Duration("remaining", w.remainingTime()),
	)
}

func (w workItem) remainingTime() time.Duration {
	if w.done == 0 || w.need <= w.done {
		return 0
	}
	avgStep := w.taken / time.Duration(w.done)
	remainingSteps := w.need - w.done
	return avgStep * time.Duration(remainingSteps)
}
