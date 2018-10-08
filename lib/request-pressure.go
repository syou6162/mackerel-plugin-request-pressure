package requestpressure

import (
	"flag"
	"fmt"
	"os"
	"time"

	"net"
	"net/http"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/montanaflynn/stats"
)

// MemoCountPlugin is mackerel plugin
type MemoCountPlugin struct {
	prefix string
	url    string
	accessNum int
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

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

// To disenable keep-alive, use custom http client
var netClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}

func durationToFetch(url string) (float64, error) {
	start := time.Now()
	resp, err := netClient.Get(url)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()
	end := time.Now()
	return end.Sub(start).Seconds(), nil
}

// FetchMetrics interface for mackerelplugin
func (p *MemoCountPlugin) FetchMetrics() (map[string]float64, error) {
	ret := make(map[string]float64)
	url := p.url
	result := make([]float64, 0)
	for i := 1; i <= p.accessNum; i++ {
		duration, _ := durationToFetch(url)
		result = append(result, duration)
	}

	average, err := stats.Mean(result)
	if err != nil {
		return nil, err
	}
	ret["average"] = average

	percentile90, err := stats.Percentile(result, 90)
	if err != nil {
		return nil, err
	}
	ret["90_percentile"] = percentile90

	percentile95, err := stats.Percentile(result, 95)
	if err != nil {
		return nil, err
	}
	ret["95_percentile"] = percentile95

	percentile99, err := stats.Percentile(result, 99)
	if err != nil {
		return nil, err
	}
	ret["99_percentile"] = percentile99
	return ret, nil
}

// Do the plugin
func Do() {
	var (
		optPrefix = flag.String("metric-key-prefix", "RequestPressure", "Metric key prefix")
		optAccessNum = flag.Int("access-num", 20, "Access number")
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
		prefix: *optPrefix,
		accessNum: *optAccessNum,
		url:    flag.Args()[0],
	}).Run()
}
