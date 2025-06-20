package main

import (
	"fmt"
	"math/rand"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/agent"
	"github.com/Morgahl/gotp/logger"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	agent, err := newAgent(baseline(100.0))
	if err != nil {
		panic(err)
	}
	count := runMissions(agent, rand.Intn(50)+50)
	value, err := agent.EvaluatePerformance()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Performance Evaluation: %f\n", value/float64(count+1))
}

func baseline[N number](base N) agent.InitFn[N] {
	if base <= 0 {
		base = 1 // Ensure a positive baseline
	}
	return func() *N {
		return &base
	}
}

func runMissions(agent *Agent[float64], numMissions int) int {
	for i := 0; i < numMissions; i++ {
		agent.LogMissionObjectives((rand.Float64() * 75) + 25)
		agent.LogMissionAttrition(rand.Float64() * 50)
		agent.LogMissionObjectives((rand.Float64() * 75) + 25)
		agent.LogMissionAttrition(rand.Float64() * 50)
	}
	return numMissions
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
	c := &Agent[N]{pid: server.PID()}
	return c, nil
}

func (c *Agent[N]) LogMissionObjectives(n N) error {
	if n <= 0 {
		return fmt.Errorf("increment value must be greater than zero")
	}
	return agent.Update(c.pid, func(state *N) {
		*state += n
	}, 0)
}

func (c *Agent[N]) LogMissionAttrition(n N) error {
	if n <= 0 {
		return fmt.Errorf("decrement value must be greater than zero")
	}
	return agent.Update(c.pid, func(state *N) {
		*state -= n
	}, 0)
}

func (c *Agent[N]) EvaluatePerformance() (n N, err error) {
	n, _, err = agent.Get[N](c.pid, gotp.PIDZero(), func(state N) N {
		return state
	}, 0)

	return
}
