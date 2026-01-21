package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/logger"
	"github.com/Morgahl/gotp/process"
)

const (
	COUNT = 10_000_000
)

func init() {
	// runtime.GOMAXPROCS(1)
	logger.ConfigFromEnv()
}

func main() {
	CPU := runtime.NumCPU()
	for workers := 1; workers <= CPU; {
		slog.Info("Running harness", "workers", workers, "cpus", CPU)
		runHarness(workers)
		if workers == 1 {
			workers = 2
		} else {
			workers += 2
		}
	}
}

func runHarness(workers int) {
	start := time.Now()
	singleWorkerAgent, err := newAgent(baseline[float64]())
	if err != nil {
		panic(err)
	}
	defer singleWorkerAgent.Stop(process.NORMAL)

	slog.Info("Running missions with agent", "pid", singleWorkerAgent.ref.PID(), "count", COUNT)
	runStart := time.Now()
	runParallelMissions(singleWorkerAgent, COUNT, workers)
	slog.Info("Completed missions", "count", COUNT, "took", time.Since(runStart))
	value := singleWorkerAgent.EvaluatePerformance()
	took := time.Since(start)
	slog.Info("Performance Evaluation", "result", value, "took", took, "avg", took/time.Duration(COUNT))
}

func baseline[N number]() agent.InitFn[state[N]] {
	return func() *state[N] {
		return &state[N]{}
	}
}

func runParallelMissions[N number](agent *Agent[N], numMissions int, numWorkers int) {
	missionsPerWorker := numMissions / numWorkers
	wg := sync.WaitGroup{}
	wg.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < missionsPerWorker; i++ {
				// slog.Debug("Worker running mission", "worker", workerID, "mission", i+1)
				score := N((rand.Float64() * 25) + 75)
				loss := N(rand.Float64() * (100 - float64(score)))
				// slog.Debug("Worker submitting report", "worker", workerID, "mission", i+1, "score", score, "loss", loss, "net", score-loss)
				agent.LogMission(score, loss)
				// slog.Debug("Worker mission report submitted", "worker", workerID, "mission", i+1)
			}
		}(w)
	}
	wg.Wait()
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type Agent[N number] struct {
	ref process.Ref
}

func newAgent[N number](initFn agent.InitFn[state[N]]) (*Agent[N], error) {
	server, err := agent.Start(initFn)
	if err != nil {
		return nil, err
	}
	c := &Agent[N]{ref: server}
	return c, nil
}

func (c *Agent[N]) Stop(reason error) {
	agent.Stop(c.ref, reason)
}

func (c *Agent[N]) LogMission(score, loss N) {
	agent.Update(c.ref, func(state *state[N]) { state.Update(score - loss) })
}

func (c *Agent[N]) EvaluatePerformance() N {
	n, _ := agent.Get(c.ref, c.ref.PID(), func(state state[N]) state[N] { return state }, 0)
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
