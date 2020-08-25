package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

const (
	namespace      = "bind"
	EXPORTER       = "bind_stats_exporter"
	RESOLVER_STATS = "resolver_stats"
	CACHE_STATS    = "cache_stats"
)

func main() {
	var (
		bindSh        = flag.String("bind.sh", "./stats.sh", "Path name of shell.")
		bindStats     = flag.String("bind.stats-file", "/var/named/data/named_stats.txt", "Path name of the status statistics file output by Bind DNS.")
		bindPidFile   = flag.String("bind.pid-file", "/run/named/named.pid", "Path to Bind's pid file to export process information.")
		showVersion   = flag.Bool("version", false, "Print version information.")
		listenAddress = flag.String("web.listen-address", ":9219", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print(EXPORTER))
		os.Exit(0)
	}
	log.Infoln("Starting", EXPORTER, version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(
		version.NewCollector(EXPORTER),
		NewStatsCollector(*bindStats, *bindSh),
	)
	if *bindPidFile != "" {
		procExporter := prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
			PidFn: func() (int, error) {
				content, err := ioutil.ReadFile(*bindPidFile)
				if err != nil {
					return 0, fmt.Errorf("Can't read pid file: %s", err)
				}
				value, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					return 0, fmt.Errorf("Can't parse pid file: %s", err)
				}
				return value, nil
			},
			Namespace: namespace,
		})
		prometheus.MustRegister(procExporter)
	}

	log.Info("Starting Server: ", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Bind Exporter</title></head>
             <body>
             <h1>Bind Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
