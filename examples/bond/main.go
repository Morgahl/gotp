package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/logger"
	"github.com/Morgahl/gotp/process"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	// runtime.GOMAXPROCS(1)
	start := time.Now()
	agent, err := newAgent(baseline[float64]())
	if err != nil {
		panic(err)
	}
	count := 10_000_000
	slog.Info("Running missions with agent", "pid", agent.pid, "count", count)
	runStart := time.Now()
	runMissions(agent, count)
	slog.Info("Completed missions", "count", count, "took", time.Since(runStart))
	value := agent.EvaluatePerformance()
	took := time.Since(start)
	slog.Info("Performance Evaluation", "result", value, "took", took, "avg", took/time.Duration(count))
}

func baseline[N number]() agent.InitFn[state[N]] {
	return func() *state[N] {
		return &state[N]{}
	}
}

func runMissions(agent *Agent[float64], numMissions int) {
	for i := 0; i < numMissions; i++ {
		// slog.Debug("Running mission", "mission", i+1)
		score := (rand.Float64() * 25) + 75
		loss := rand.Float64() * (100 - score)
		// slog.Debug("Submitting report", "mission", i+1, "score", score, "loss", loss, "net", score-loss)
		agent.LogMission(score, loss)
		// slog.Debug("Mission report submitted", "mission", i+1)
	}
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type Agent[N number] struct {
	pid process.PID
}

func newAgent[N number](initFn agent.InitFn[state[N]]) (*Agent[N], error) {
	server, err := agent.Start(initFn)
	if err != nil {
		return nil, err
	}
	c := &Agent[N]{pid: server.PID()}
	return c, nil
}

func (c *Agent[N]) LogMission(score, loss N) {
	agent.Update(c.pid, func(state *state[N]) { state.Update(score - loss) })
}

func (c *Agent[N]) EvaluatePerformance() N {
	n, _ := agent.Get(c.pid, c.pid, func(state state[N]) state[N] { return state }, 0)
	slog.Info("Evaluating performance", "state", n)
	return n.EvaluatePerformance()
}

type state[N number] struct {
	value N
	count N
}

func (s *state[N]) Update(n N) {
	s.value += n
	s.count += 1
}

func (s *state[N]) EvaluatePerformance() N {
	return s.value / s.count
}

func (s state[N]) String() string {
	return fmt.Sprintf("State{value: %v, count: %v}", s.value, s.count)
}

func (s state[N]) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("value", s.value),
		slog.Any("count", s.count),
	)
}
