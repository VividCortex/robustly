// Package robustly provides code to handle (and create) infrequent panics.
package robustly

// Copyright (c) 2013 VividCortex, Inc. All rights reserved.
// Please see the LICENSE file for applicable license terms.

import (
	"fmt"
	"github.com/VividCortex/ewma"
	"log"
	"runtime/debug"
	"time"
)

// RunOptions is a struct to hold the optional arguments to Run.
type RunOptions struct {
	RateLimit  float64       // the rate limit in crashes per second
	Timeout    time.Duration // the timeout (after which Run will stop trying)
	PrintStack bool          // whether to print the panic stacktrace or not
	RetryDelay time.Duration // inject a delay before retrying the run
}

// Run runs the given function robustly, catching and restarting on panics.
// Takes a RunOptions struct pointer as options, nil to use the default parameters.
//
//  Default parameters are:
//  RateLimit: 1.0,                 // the rate limit in crashes per second
//  Timeout: 1 * time.Second,       // the timeout (after which Run will stop trying)
//  PrintStack: false,              // whether to print the panic stacktrace or not
//  RetryDelay: 0 * time.Nanosecond // inject a delay before retrying the run
func Run(function func(), options *RunOptions) int {
	if options == nil {
		options = &RunOptions{
			RateLimit:  1.0,
			Timeout:    1 * time.Second,
			PrintStack: false,
			RetryDelay: 0 * time.Nanosecond,
		}
	}
	if options.RateLimit == 0 {
		log.Print("[robustly] warning: RateLimit is 0")
	}
	// We use a moving average to compute the rate of errors per second.
	avg := ewma.NewMovingAverage(options.Timeout.Seconds())
	before := time.Now()
	var startAboveLimit time.Time
	var belowLimit bool = true
	var beforeTimeout = true
	var totalPanics = 0
	var oktorun bool = true

	for oktorun {
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
					if avg.Value() > options.RateLimit {
						if belowLimit {
							startAboveLimit = after
						}
						beforeTimeout =
							after.Before(startAboveLimit.Add(options.Timeout))
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

				if options.PrintStack {
					log.Printf("[robustly] %v\n%s\n", localErr, debug.Stack())
				}

				if options.RetryDelay > time.Nanosecond*0 {
					time.Sleep(options.RetryDelay)
				}
			}()
			function()
			return
		}()

	}
	return totalPanics
}
