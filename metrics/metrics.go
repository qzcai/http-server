package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// ExecutionTimer measures execution time of a computation, split into major steps
// usual usage pattern is: timer := NewExecutionTimer(...) ; compute ; timer.ObserveStep() ; ... ; timer.ObserveTotal()
type ExecutionTimer struct {
	histo *prometheus.HistogramVec
	start time.Time
	last  time.Time
}

const (
	Namespace = "httpserver"
)

var (
	latency = CreateExecutionTimeMetric(Namespace, "Time spent.")
)

func Register() {
	if err := prometheus.Register(latency); err != nil {
		fmt.Println(err)
	}
}

// NewExecutionTimer provides a timer for admission latency; call ObserveXXX() on it to measure
func NewExecutionTimer() *ExecutionTimer {
	now := time.Now()
	return &ExecutionTimer{
		histo: latency,
		start: now,
		last:  now,
	}
}

// ObserveTotal measures the execution time from the creation of the ExecutionTimer
func (t *ExecutionTimer) ObserveTotal() {
	(*t.histo).WithLabelValues("total").Observe(time.Now().Sub(t.start).Seconds())
}

// CreateExecutionTimeMetric prepares a new histogram labeled with execution step
func CreateExecutionTimeMetric(namespace string, help string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "execution_latency_seconds",
			Help:      help,
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		}, []string{"step"},
	)
}