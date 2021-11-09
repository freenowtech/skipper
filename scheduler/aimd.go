package scheduler

import (
	"math"
	"sync"
	"time"

	"github.com/zalando/skipper/metrics"
)

const (
	threshold = 3
)

type AIMD struct {
	count    float64
	duration float64
	mu       sync.Mutex
	metrics  metrics.Metrics
	interval time.Duration

	// minConcurrency is the minimum allowed number of concurrent connections
	minConcurrency float64
	// maxConcurrency is the maximum allowed number of concurrent connections
	maxConcurrency float64
}

func NewAIMD(minConcurrency, maxConcurrency float64, interval time.Duration) *AIMD {
	return &AIMD{
		minConcurrency: minConcurrency,
		maxConcurrency: maxConcurrency,
		interval:       interval,
	}
}

func (a *AIMD) Reset() {
	a.mu.Lock()
	a.count = 0
	a.duration = 0
	a.mu.Unlock()
}

func (a *AIMD) Total() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.duration / a.count
}

func (a *AIMD) Collect(duration float64) {
	a.mu.Lock()
	a.count++
	a.duration += duration
	a.mu.Unlock()
}

func (a *AIMD) CalculateRPS() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.count / a.interval.Seconds()
}

func (a *AIMD) CalculateConcurrency(oldConcurrency float64) int {
	rps := a.CalculateRPS()
	if rps < 1 {
		return int(a.maxConcurrency)
	}

	if a.Total() < threshold {
		return int(math.Min(oldConcurrency+1, a.maxConcurrency))
	}
	return int(math.Max(oldConcurrency/2, a.minConcurrency))
}

func SetConcurrency(q *Queue) {
	t := time.NewTicker(q.aimd.interval)
	for {
		<-t.C
		newConcurrency := q.aimd.CalculateConcurrency(float64(q.config.MaxConcurrency))
		if q.metrics != nil {
			q.metrics.UpdateGauge(q.aimdConcurrencyMetricsKey, float64(newConcurrency))
		}

		q.config.MaxConcurrency = newConcurrency
		q.reconfigure()
		q.aimd.Reset()
	}
}
