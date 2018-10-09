package requestpressure

import (
	"flag"
	"fmt"
	"os"
	"time"

	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	vegeta "github.com/tsenart/vegeta/lib"
)

// MemoCountPlugin is mackerel plugin
type MemoCountPlugin struct {
	prefix      string
	url         string
	accessNum   int
	durationSec int
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p *MemoCountPlugin) MetricKeyPrefix() string {
	return p.prefix
}

// GraphDefinition interface for mackerelplugin
func (p *MemoCountPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(p.prefix)
	labelPrefix = strings.Replace(labelPrefix, "-", " ", -1)
	labelPrefix = strings.Replace(labelPrefix, "RequestPressure", "RequestPressure", -1)
	return map[string]mp.Graphs{
		"latency": {
			Label: labelPrefix + " Latency",
			Unit:  "float",
			Metrics: []mp.Metrics{
				{Name: "average", Label: "Average"},
				{Name: "90_percentile", Label: "90 Percentile"},
				{Name: "95_percentile", Label: "95 Percentile"},
				{Name: "99_percentile", Label: "99 Percentile"},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p *MemoCountPlugin) FetchMetrics() (map[string]float64, error) {
	ret := make(map[string]float64)
	url := p.url

	rate := vegeta.Rate{Freq: p.accessNum, Per: time.Second}
	duration := time.Duration(p.durationSec) * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    url,
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	average := metrics.Latencies.Mean.Seconds()
	ret["average"] = average

	percentile90 := metrics.Latencies.Quantile(0.90).Seconds()
	ret["90_percentile"] = percentile90

	percentile95 := metrics.Latencies.P95.Seconds()
	ret["95_percentile"] = percentile95

	percentile99 := metrics.Latencies.P99.Seconds()
	ret["99_percentile"] = percentile99
	return ret, nil
}

// Do the plugin
func Do() {
	var (
		optPrefix      = flag.String("metric-key-prefix", "RequestPressure", "Metric key prefix")
		optAccessNum   = flag.Int("access-num", 2, "Access number")
		optDurationSec = flag.Int("duration", 10, "duration seconds")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION] url\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	mp.NewMackerelPlugin(&MemoCountPlugin{
		prefix:      *optPrefix,
		accessNum:   *optAccessNum,
		durationSec: *optDurationSec,
		url:         flag.Args()[0],
	}).Run()
}
