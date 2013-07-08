package robustly

import (
	"testing"
	"time"
)

//func init() {
//	log.SetLevel(log.Ldisabled)
//}

func TestRobustly1(t *testing.T) {
	tries := 0
	options := map[string]int {
		"rateLimit": 1,
		"timeout": 1,
	}

	panics := Run(func() { PanicRateIters(time.Second, 5, &tries) }, options)
	if panics != 5 {
		t.Errorf("function panicked %d times, expected 5", panics)
	}
}

func TestRobustly2(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	}()
	options := map[string]int {
		"rateLimit": 1,
		"timeout": 1,
	}
	tries := 0
	Run(func() { PanicRateIters(time.Millisecond*300, 500, &tries) }, options)
	t.Errorf("this code shouldn't run at all, the defer() should run")
}

func TestRobustly3(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Errorf("got error %v, expected nil", err)
		}
	}()
	options := map[string]int {
		"rateLimit": 1,
		"timeout": 1,
	}
	tries := 0
	panics := Run(func() { PanicRateIters(time.Millisecond*300, 2, &tries) }, options)
	if panics != 2 {
		t.Errorf("got %d panics, expected 2", panics)
	}
}

func PanicRateIters(rate time.Duration, iters int, count *int) {
	time.Sleep(rate)
	*count = *count + 1
	if *count <= iters {
		Panic()
	}
}

func Panic() {
	sl := make([]int, 0)
	sl[1] = 1 // index range panic
}
