package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"


	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// With certain versions of Kingpin, if flags are not in the main package they dont get processes correctly.
var (
	maxScrapeTimeout     = kingpin.Flag("scrape.max-timeout", "Any scrape with a timeout higher than this will have to be clamped to this.").Default("5m").Duration()
	defaultScrapeTimeout = kingpin.Flag("scrape.default-timeout", "If a scrape lacks a timeout, use this value.").Default("15s").Duration()
)

func GetScrapeTimeout(h http.Header) time.Duration {
	return GetScrapeTimeoutWithLogger(h, nil)
}

func GetScrapeTimeoutWithLogger(h http.Header, logger log.Logger) time.Duration {

	timeout := *defaultScrapeTimeout
	timeoutSeconds, err := strconv.ParseFloat(h.Get("X-Prometheus-Scrape-Timeout-Seconds"), 64)
	if  logger != nil {
		level.Error(logger).Log("msg", "Parsing Timeout ", "timeout",  timeoutSeconds, "err", err)
	}
	if err == nil {
		timeout = time.Duration(timeoutSeconds * 1e9)
		if  logger != nil {
			level.Error(logger).Log("msg", "Accepted Timeout ", "to", timeout)
		}
	} else {
		if  logger != nil {
			level.Error(logger).Log("msg", "Timeout not accepted ")
		}
	}
	if timeout > *maxScrapeTimeout {
		timeout = *maxScrapeTimeout
		if  logger != nil {
			level.Error(logger).Log("msg", "Over Max, defaulting to ", "to", timeout)
		}
	}
	if  logger != nil {
			level.Error(logger).Log("msg", "Final Timeout ", "to", timeout)
	}
	return timeout
}
