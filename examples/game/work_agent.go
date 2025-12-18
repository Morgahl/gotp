package game

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

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
			return &s
		},
		process.Named(w.name),
	)
}

func ContextGetWork[S process.Sendable](pctx process.Context, agnt S, from process.PID) (*workItem, bool) {
	s, ok := agent.ContextGetAndUpdate(pctx, agnt, func(state *state) state {
		state.generate()
		return *state
	}, 5*time.Second)
	if !ok || !s.next {
		return nil, false
	}
	return s.workItem, s.next
}

func ContextSubmitProcessedWork[S process.Sendable](pctx process.Context, agnt S, work *workItem) (*workItem, bool) {
	s, ok := agent.ContextGetAndUpdate(pctx, agnt, func(state *state) state {
		if state.receivedProcessed(work) {
			slog.WarnContext(pctx.Context(), "Submitted work", slog.Uint64("wanted", state.wanted), slog.Uint64("generated", state.generated), slog.Uint64("processed", state.processed))
		}
		if state.processed == state.wanted {
			slog.InfoContext(pctx.Context(), "All work processed", slog.Uint64("wanted", state.wanted), slog.Uint64("generated", state.generated), slog.Uint64("processed", state.processed))
			agent.ContextStop(pctx, agnt, process.NORMAL)
		}
		return *state
	}, 5*time.Second)
	if !ok || !s.next {
		return nil, false
	}
	return s.workItem, s.next
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
		need: uint64(rand.Intn(56) + 5),
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

func (s *state) stepNextLog() {
	if s.wanted == 0 || s.processed >= s.wanted {
		s.nextLog = s.wanted + 1
		return
	}
	base := s.processed
	pos := float64(s.processed) / float64(s.wanted)
	if pos < 0 {
		pos = 0
	} else if pos > 1 {
		pos = 1
	}
	maxI := float64(s.wanted) * 0.01
	minI := 1.0
	easeIn := math.Pow(pos, 4.0)
	interval := maxI - (maxI-minI)*easeIn
	step := uint64(math.Max(1, math.Round(interval)))
	s.nextLog = base + step
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
