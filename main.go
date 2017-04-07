package county

import "time"

type incrementOp struct {
    key string
    resp chan int
}

type decrementOp struct {
    key string
    resp chan int
}

type statsOp struct {
    resp chan map[string]int
}

type Counter struct {
    increment chan *incrementOp
    decrement chan *decrementOp
    stats     chan *statsOp
}

func (c *Counter) Increment(key string) int {
    resp := make(chan int)
    inc := &incrementOp{key, resp}
    c.increment <- inc
    return <-resp
}

func (c *Counter) Decrement(key string) int {
    resp := make(chan int)
    dec := &decrementOp{key, resp}
    c.decrement <- dec
    return <-resp
}

func (c *Counter) Stats() map[string] int {
    resp := make(chan map[string]int)
    stats := &statsOp{resp}
    c.stats <- stats
    return <-resp
}

func NewCounter(expiration time.Duration) *Counter {
    increments := make(chan *incrementOp)
    decrements := make(chan *decrementOp)
    stats      := make(chan *statsOp)
    counters   := map[string]int{}
    go func() {
        for {
            select {
            case inc := <-increments:
                counters[inc.key] += 1
                inc.resp <- counters[inc.key]
                go func() {
                    time.Sleep(expiration)
                    op := &decrementOp{inc.key, make(chan int)}
                    decrements <- op
                    <-op.resp
                }()
            case dec := <-decrements:
                counters[dec.key] -= 1
                value := counters[dec.key]
                if value == 0 {
                    delete(counters, dec.key)
                }
                dec.resp <- value
            case stat := <-stats:
                currentData := map[string]int{}
                for k, v := range counters {
                    currentData[k] = v
                }
                stat.resp <- currentData
            }
        }
    }()
    return &Counter{increments, decrements, stats}
}
