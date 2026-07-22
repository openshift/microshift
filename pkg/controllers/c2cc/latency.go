package c2cc

import (
	"math"
	"time"

	microshiftv1alpha1 "github.com/openshift/microshift/pkg/apis/microshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const windowSize = 20

type latencyWindow struct {
	samples [windowSize]time.Duration
	pos     int
	count   int
}

func (w *latencyWindow) add(d time.Duration) {
	w.samples[w.pos] = d
	w.pos = (w.pos + 1) % windowSize
	if w.count < windowSize {
		w.count++
	}
}

func (w *latencyWindow) stats() *microshiftv1alpha1.LatencyStats {
	if w.count == 0 {
		return nil
	}

	n := w.count
	start := 0
	if n == windowSize {
		start = w.pos
	}

	var sum time.Duration
	minD := w.samples[start]
	maxD := w.samples[start]
	var last time.Duration

	for i := 0; i < n; i++ {
		idx := (start + i) % windowSize
		d := w.samples[idx]
		sum += d
		if d < minD {
			minD = d
		}
		if d > maxD {
			maxD = d
		}
		last = d
	}

	avg := sum / time.Duration(n)

	var varianceSum float64
	avgF := float64(avg)
	for i := 0; i < n; i++ {
		idx := (start + i) % windowSize
		diff := float64(w.samples[idx]) - avgF
		varianceSum += diff * diff
	}
	stddev := time.Duration(math.Sqrt(varianceSum / float64(n)))

	return &microshiftv1alpha1.LatencyStats{
		Avg:    metav1.Duration{Duration: avg},
		Min:    metav1.Duration{Duration: minD},
		Max:    metav1.Duration{Duration: maxD},
		Last:   metav1.Duration{Duration: last},
		Stddev: metav1.Duration{Duration: stddev},
	}
}
