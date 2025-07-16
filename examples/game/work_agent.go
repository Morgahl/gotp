package game

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	gotp "github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

func WorkAgent(name gotp.Atom, wanted uint64) server.Supervisable {
	return agent.New(
		func() *state { return &state{wanted: wanted} },
		process.Named(name),
	)
}

func GetWork(name gotp.Atom, from process.PID) (workItem, bool) {
	s, ok := agent.GetAndUpdate(name, from, func(state *state) state {
		state.generate()
		return *state
	})
	if !ok || !s.next {
		return workItem{}, false
	}
	return s.workItem, s.next
}

func SubmitProcessedWork[S process.Sendable](agnt S, work workItem) {
	agent.Update(agnt, func(state *state) {
		slog.Info("Submitted work", "work", work)
		state.receivedProcessed(work)
		if state.processed == state.wanted {
			slog.Info("All work processed", "generated", state.generated, "processed", state.processed, "wanted", state.wanted)
			agent.Stop(agnt, process.NORMAL)
		}
	})
}

type state struct {
	wanted, generated, processed uint64
	next                         bool
	workItem                     workItem
}

func (s *state) generate() {
	if s.generated >= s.wanted {
		s.workItem = workItem{}
		s.next = false
		return
	}
	s.generated++
	s.next = true
	s.workItem = workItem{
		id:   gotp.Atom(fmt.Sprintf("work-%d", s.generated)),
		need: uint64(rand.Intn(56) + 5),
	}
}

func (s *state) receivedProcessed(work workItem) {
	if work.need == work.done {
		s.processed++
	}
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
