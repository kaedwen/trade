package app

import (
	"context"
	"time"

	"github.com/kaedwen/trade/pkg/base"
	"github.com/procyon-projects/chrono"
)

var checkQuoteTask *chrono.ScheduledTask

func checkQuote(targets []string, quoteChannel chan base.Quote) {
	for _, target := range targets {
		if _, err := base.GetQuote(target); err == nil {
			base.FatalIfError(err)
			// quoteChannel <- quote
		}
	}
}

func SetupCheckQuote(rate time.Duration, delay time.Duration, quoteChannel chan base.Quote) {
	task, err := taskScheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		checkQuote(base.Config.Targets, quoteChannel)
	}, rate, chrono.WithTime(time.Now().Add(delay)))
	base.FatalIfError(err)

	checkQuoteTask = &task

}

func StopCheckQuote() {
	if checkQuoteTask != nil {
		(*checkQuoteTask).Cancel()
		checkQuoteTask = nil
	}
}
