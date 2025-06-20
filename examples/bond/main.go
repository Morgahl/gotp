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
	logger.SetGlobalDefaultLogger()
}

func main() {
	counter, err := newCounterAgent[uint8](gen)
	if err != nil {
		panic(err)
	}
	counter.Increment(uint8((rand.Intn(102))) + 25) // Increment by a random value between 25 and 100
	counter.Increment(uint8((rand.Intn(102))) + 25)
	counter.Decrement(uint8((rand.Intn(50))))
	value, ok, err := counter.Get()
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("Counter value not found")
	}
	fmt.Println("Counter value:", value)
}
func gen[N integer]() *N {
	var n N
	return &n
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Counter[N integer] struct {
	pid gotp.PID
}

func newCounterAgent[N integer](initFn agent.InitFn[N]) (*Counter[N], error) {
	server, err := agent.New(initFn).Start()
	if err != nil {
		return nil, err
	}
	c := &Counter[N]{pid: server.PID()}
	return c, nil
}

func (c *Counter[N]) Increment(n N) error {
	if n <= 0 {
		return fmt.Errorf("increment value must be greater than zero")
	}
	return agent.Update(c.pid, func(state *N) {
		*state += n
	}, 0)
}

func (c *Counter[N]) Decrement(n N) error {
	if n <= 0 {
		return fmt.Errorf("decrement value must be greater than zero")
	}
	return agent.Update(c.pid, func(state *N) {
		*state -= n
	}, 0)
}

func (c *Counter[N]) Get() (N, bool, error) {
	return agent.Get(c.pid, gotp.PIDZero(), func(state N) N {
		return state
	}, 0)
}
