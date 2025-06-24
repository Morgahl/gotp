package main

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/logger"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	start := time.Now()
	agent, err := newAgent(baseline(100.0))
	if err != nil {
		panic(err)
	}
	slog.Info("Running missions with agent", "pid", agent.pid)
	count := rand.Intn(10) + 5
	runMissions(agent, count)
	slog.Info("Completed missions", "count", count)
	value := agent.EvaluatePerformance()
	slog.Info("Performance Evaluation", "result", value/float64(count+1), "took", time.Since(start))
}

func baseline[N number](base N) agent.InitFn[N] {
	if base <= 0 {
		base = 1
	}
	return func() *N {
		return &base
	}
}

func runMissions(agent *Agent[float64], numMissions int) {
	for i := 0; i < numMissions; i++ {
		// slog.Info("Running mission", "mission", i+1)
		score := (rand.Float64() * 25) + 75
		loss := rand.Float64() * 25
		// slog.Info("Submitting report", "mission", i+1, "score", score, "loss", loss, "net", score-loss)
		agent.LogMissionObjectives(score)
		agent.LogMissionAttrition(loss)
		// slog.Info("Mission report submitted", "mission", i+1)
	}
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type Agent[N number] struct {
	pid gotp.PID
}

func newAgent[N number](initFn agent.InitFn[N]) (*Agent[N], error) {
	server, err := agent.New(initFn).Start()
	if err != nil {
		return nil, err
	}
	c := &Agent[N]{pid: server.ID()}
	return c, nil
}

func (c *Agent[N]) LogMissionObjectives(n N) {
	if n < 0 {
		n = 0
	}
	agent.Update(c.pid, func(state *N) {
		*state += n
		slog.Debug("Logged mission objectives result", "result", n, "state", *state)
	})
}

func (c *Agent[N]) LogMissionAttrition(n N) {
	if n < 0 {
		n = 0
	}
	agent.Update(c.pid, func(state *N) {
		*state -= n
		slog.Debug("Logged mission attrition result", "result", n, "state", *state)
	})
}

func (c *Agent[N]) EvaluatePerformance() (n N) {
	n, _ = agent.Get(c.pid, c.pid, func(state N) N {
		return state
	})
	return
}
