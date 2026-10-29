package transport

import (
	"math"
	"testing"
	"time"
)

func TestRetryDelay(t *testing.T) {
	minimum := 250 * time.Millisecond
	for _, testCase := range []struct {
		name    string
		seconds float64
		want    time.Duration
		valid   bool
	}{
		{name: "negative", seconds: -1, want: minimum, valid: true},
		{name: "zero", seconds: 0, want: minimum, valid: true},
		{name: "below minimum", seconds: 0.1, want: minimum, valid: true},
		{name: "fractional seconds", seconds: 1.5, want: 1500 * time.Millisecond, valid: true},
		{name: "not a number", seconds: math.NaN()},
		{name: "positive infinity", seconds: math.Inf(1)},
		{name: "negative infinity", seconds: math.Inf(-1)},
		{name: "duration overflow", seconds: float64(math.MaxInt64) / float64(time.Second)},
	} {
		t.Run(testCase.name, func(subtest *testing.T) {
			actual, valid := RetryDelay(minimum, &testCase.seconds)
			if actual != testCase.want || valid != testCase.valid {
				subtest.Errorf("RetryDelay(%v, %v) = (%v, %t), want (%v, %t)", minimum, testCase.seconds, actual, valid, testCase.want, testCase.valid)
			}
		})
	}
	if actual, valid := RetryDelay(minimum, nil); actual != minimum || !valid {
		t.Errorf("RetryDelay(%v, nil) = (%v, %t), want (%v, true)", minimum, actual, valid, minimum)
	}
}
