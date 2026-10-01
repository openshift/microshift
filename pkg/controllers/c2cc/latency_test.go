package c2cc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatencyWindow_EmptyReturnsNil(t *testing.T) {
	w := &latencyWindow{}
	assert.Nil(t, w.stats())
}

func TestLatencyWindow_SingleSample(t *testing.T) {
	w := &latencyWindow{}
	w.add(10 * time.Millisecond)

	s := w.stats()
	require.NotNil(t, s)
	assert.Equal(t, 10*time.Millisecond, s.Avg.Duration)
	assert.Equal(t, 10*time.Millisecond, s.Min.Duration)
	assert.Equal(t, 10*time.Millisecond, s.Max.Duration)
	assert.Equal(t, 10*time.Millisecond, s.Last.Duration)
	assert.Equal(t, time.Duration(0), s.Stddev.Duration)
}

func TestLatencyWindow_PartialFill(t *testing.T) {
	w := &latencyWindow{}
	w.add(10 * time.Millisecond)
	w.add(20 * time.Millisecond)
	w.add(30 * time.Millisecond)

	s := w.stats()
	require.NotNil(t, s)
	assert.Equal(t, 20*time.Millisecond, s.Avg.Duration)
	assert.Equal(t, 10*time.Millisecond, s.Min.Duration)
	assert.Equal(t, 30*time.Millisecond, s.Max.Duration)
	assert.Equal(t, 30*time.Millisecond, s.Last.Duration)
	assert.True(t, s.Stddev.Duration > 0, "stddev should be > 0 for varied samples")
}

func TestLatencyWindow_FullWindowWraps(t *testing.T) {
	w := &latencyWindow{}

	// Fill with 25 samples: values 1ms through 25ms
	for i := 1; i <= 25; i++ {
		w.add(time.Duration(i) * time.Millisecond)
	}

	assert.Equal(t, windowSize, w.count)

	s := w.stats()
	require.NotNil(t, s)

	// Last 20 samples: 6ms through 25ms
	assert.Equal(t, 25*time.Millisecond, s.Last.Duration)
	assert.Equal(t, 6*time.Millisecond, s.Min.Duration)
	assert.Equal(t, 25*time.Millisecond, s.Max.Duration)

	// Avg of 6..25 = (6+25)/2 = 15.5ms, truncated to 15ms by integer division
	expectedAvg := time.Duration(15) * time.Millisecond
	assert.InDelta(t, float64(expectedAvg), float64(s.Avg.Duration), float64(time.Millisecond))
}

func TestLatencyWindow_StatsComputation(t *testing.T) {
	w := &latencyWindow{}
	// 5 samples: 100, 200, 300, 400, 500 µs
	for i := 1; i <= 5; i++ {
		w.add(time.Duration(i*100) * time.Microsecond)
	}

	s := w.stats()
	require.NotNil(t, s)

	// avg = 300µs
	assert.Equal(t, 300*time.Microsecond, s.Avg.Duration)
	assert.Equal(t, 100*time.Microsecond, s.Min.Duration)
	assert.Equal(t, 500*time.Microsecond, s.Max.Duration)
	assert.Equal(t, 500*time.Microsecond, s.Last.Duration)

	// stddev = sqrt(mean((x-300)^2)) = sqrt((40000+10000+0+10000+40000)/5) = sqrt(20000) ≈ 141.4µs
	assert.InDelta(t, float64(141*time.Microsecond), float64(s.Stddev.Duration), float64(2*time.Microsecond))
}

func TestLatencyWindow_IdenticalSamples(t *testing.T) {
	w := &latencyWindow{}
	for i := 0; i < 10; i++ {
		w.add(5 * time.Millisecond)
	}

	s := w.stats()
	require.NotNil(t, s)
	assert.Equal(t, 5*time.Millisecond, s.Avg.Duration)
	assert.Equal(t, 5*time.Millisecond, s.Min.Duration)
	assert.Equal(t, 5*time.Millisecond, s.Max.Duration)
	assert.Equal(t, 5*time.Millisecond, s.Last.Duration)
	assert.Equal(t, time.Duration(0), s.Stddev.Duration)
}
