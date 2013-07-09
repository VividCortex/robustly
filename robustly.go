package robustly
// Package robustly provides code to handle infrequent panics.

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"time"
	"github.com/VividCortex/moving_average"
)

func Run(function func(), options map[string]int) int {
	// Set config options
	rateLimit := options["rateLimit"]
	timeout := options["timeout"]

	// We use a moving average to compute the rate of errors per second.
	avg := moving_average.NewMovingAverage(float64(timeout))
	before := time.Now()
	var startAboveLimit time.Time
	var belowLimit bool = true
	var beforeTimeout = true
	var totalPanics = 0
	var oktorun bool = true

	for oktorun {
		fmt.Printf("running function resiliently, panic rate %f since %s \n",
								avg.Value(),
								startAboveLimit,
								)

		// Run the provided code and catch errors.
		func() {
			defer func() {
				localErr := recover()
				if localErr == nil {
					oktorun = false // The call to f() exited normally.
					return
				}

				totalPanics++
				after := time.Now()
				duration := after.Sub(before).Seconds()
				if duration > 0 {
					rate := 1.0 / duration
					avg.Add(rate)

					// Figure out whether we're above the rate limit and for how long
					if avg.Value() > float64(rateLimit) {
						if belowLimit {
							startAboveLimit = after
						}
						beforeTimeout =
							after.Before(startAboveLimit.Add(time.Second * time.Duration(timeout)))
						belowLimit = false
					} else {
						belowLimit = true
					}
				}
				before = after

				if !belowLimit && !beforeTimeout {
					panic(fmt.Sprintf("giving up after %d errors at %.2f/sec since %s",
						totalPanics, avg.Value(), startAboveLimit))
				}

				var buf bytes.Buffer
				fmt.Fprintf(&buf, "%v\n", localErr)
				buf.Write(debug.Stack())
				//fmt.Printf(buf.String())
			}()
			function()
			return
		}()

	}
	return totalPanics
}
