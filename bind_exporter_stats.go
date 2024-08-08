package main

import (
	"bytes"
	"io/ioutil"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// ftp://ftp.isc.org/isc/bind9/9.9.0/doc/arm/Bv9ARM.ch06.html
var (
	subReg, _  = regexp.Compile(" ?\\+\\+ ?")
	viewReg, _ = regexp.Compile("[\\(|\\)|\\<|/]")
	numReg, _  = regexp.Compile(`[0-9]+`)
	letReg, _  = regexp.Compile(`[0-9a-zA-Z()><-]+`)
	up         = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the Bind instance query successful?",
		nil, nil,
	)
	bootTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "boot_time_seconds"),
		"Start time of the BIND process since unix epoch in seconds.",
		nil, nil,
	)
	nameServerStatistics = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "name_server_stats_total"),
		"Name Server Statistics Counters.",
		[]string{"type"}, nil,
	)
	outgoingQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "outgoing_queries_total"),
		"Outgoing Queries.",
		[]string{"view", "type"}, nil,
	)
	incomingQueries = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "incoming_queries_total"),
		"Number of incoming DNS queries.",
		[]string{"type"}, nil,
	)
	incomingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "incoming_requests_total"),
		"Number of incoming DNS requests.",
		[]string{"opcode"}, nil,
	)
	resolverMetricStatsFile = map[string]*prometheus.Desc{
		"GlueFetchv4": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv4_ns_total"),
			"IPv4 NS address fetches.",
			[]string{"view"}, nil,
		),
		"GlueFetchv6": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv6_ns_total"),
			"IPv6 NS address fetches.",
			[]string{"view"}, nil,
		),
		"EDNS0Fail": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "query_edns_failures_total"),
			"EDNS(0) query failures.",
			[]string{"view"}, nil,
		),
		"Mismatch": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "response_mismatch_total"),
			"Number of mismatch responses received.",
			[]string{"view"}, nil,
		),
		"Retry": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "query_retries_total"),
			"Number of resolver query retries.",
			[]string{"view"}, nil,
		),
		"Truncated": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "response_truncated_total"),
			"Number of truncated responses received.",
			[]string{"view"}, nil,
		),
		"Queryv4": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv4_queries_sent_total"),
			"IPv4 queries sent.",
			[]string{"view"}, nil,
		),
		"Queryv6": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv6_queries_sent_total"),
			"IPv6 queries sent.",
			[]string{"view"}, nil,
		),
		"Responsev4": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv4_responses_received_total"),
			"IPv4 responses received.",
			[]string{"view"}, nil,
		),
		"Responsev6": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "ipv6_responses_received_total"),
			"IPv6 responses received.",
			[]string{"view"}, nil,
		),
		"NXDOMAIN": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "nxdomain_received_total"),
			"NXDOMAIN received.",
			[]string{"view"}, nil,
		),
		"SERVFAIL": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "servfail_received_total"),
			"SERVFAIL received.",
			[]string{"view"}, nil,
		),
		"QryRTTnn": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "queries_with_rtt_milliseconds_histogram"),
			"Frequency table on round trip times (RTTs) of queries. Each nn specifies the corresponding frequency.",
			[]string{"view"}, nil,
		),
		"QueryTimeout": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, RESOLVER_STATS, "query_timeouts_total"),
			"Query timeouts.",
			[]string{"view"}, nil,
		),
	}
	socketIO = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "socket_io_total"),
		"Socket I/O statistics counters are defined per socket types.",
		[]string{"type"}, nil,
	)
	zoneMetricStats = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "zone_maintenance_total"),
		"Zone Maintenance Statistics Counters.",
		[]string{"type"}, nil,
	)
	cacheRRsetsStats = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, CACHE_STATS, "cache_rrsets"),
		"Number of RRSets in Cache database.",
		[]string{"view", "type"}, nil,
	)
	cacheStatistics = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, CACHE_STATS, "statistics"),
		"Cache Statistics.",
		[]string{"view", "type"}, nil,
	)
	cacheMetricStatsFile = map[string]*prometheus.Desc{
		"Buckets": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "database_buckets"),
			"cache database hash buckets",
			[]string{"view"}, nil,
		),
		"Nodes": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "database_nodes"),
			"cache database nodes",
			[]string{"view"}, nil,
		),
		"UseHeapHighest": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "use_heap_highest"),
			"cache heap highest memory in use",
			[]string{"view"}, nil,
		),
		"UseHeapMemory": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "use_heap_memory"),
			"cache heap memory in use",
			[]string{"view"}, nil,
		),
		"TotalHeapMemory": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "heap_memory_total"),
			"cache heap memory total",
			[]string{"view"}, nil,
		),
		"Hits": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "hits"),
			"cache hits",
			[]string{"view"}, nil,
		),
		"QueryHits": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "query_hits"),
			"cache hits from query",
			[]string{"view"}, nil,
		),
		"Misses": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "misses"),
			"cache misses",
			[]string{"view"}, nil,
		),
		"QueryMisses": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "query_misses"),
			"cache misses from query",
			[]string{"view"}, nil,
		),
		"DelTTL": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "delete_ttl"),
			"cache records deleted due to TTL expiration",
			[]string{"view"}, nil,
		),
		"DelMem": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "delete_memory"),
			"cache records deleted due to memory exhaustion",
			[]string{"view"}, nil,
		),
		"UseTreeHighest": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "use_tree_highest"),
			"cache tree highest memory in use",
			[]string{"view"}, nil,
		),
		"UseTreeMemory": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "use_tree_memory"),
			"cache tree memory in use",
			[]string{"view"}, nil,
		),
		"TotalTreeMemory": prometheus.NewDesc(
			prometheus.BuildFQName(namespace, CACHE_STATS, "tree_memory_total"),
			"cache tree memory total",
			[]string{"view"}, nil,
		),
	}
	nameServerMap = map[string]string{
		"IPv requests received":                        "IPv",
		"TCP requests received":                        "ReqTCP",
		"duplicate queries received":                   "QryDuplicate",
		"queries caused recursion":                     "QryRecursion",
		"queries dropped":                              "QryDropped",
		"queries resulted in NXDOMAIN":                 "QryNXDOMAIN",
		"queries resulted in SERVFAIL":                 "QrySERVFAIL",
		"queries resulted in authoritative answer":     "QryAuthAns",
		"queries resulted in non authoritative answer": "QryNoauthAns",
		"queries resulted in nxrrset":                  "QryNxrrset",
		"queries resulted in referral answer":          "QryReferral",
		"queries resulted in successful answer":        "QrySuccess",
		"requests with EDNS received":                  "ReqEdns",
		"requests with TSIG received":                  "ReqTSIG",
		"responses sent":                               "Response",
		"responses with EDNS sent":                     "RespEDNS",
		"truncated responses sent":                     "RespTruncated",
	}
	resolverStatisticsMap = map[string]string{
		"IPv4 queries sent":            "Queryv4",
		"IPv6 queries sent":            "Queryv6",
		"IPv4 responses received":      "Responsev4",
		"IPv6 responses received":      "Responsev6",
		"NXDOMAIN received":            "NXDOMAIN",
		"SERVFAIL received":            "SERVFAIL",
		"FORMERR received":             "FORMERR",
		"Mismatch responses received":  "Mismatch",
		"Other errors received":        "OtherError",
		"EDNS(0) query failures":       "EDNS0Fail",
		"IPv4 NS address fetches":      "GlueFetchv4",
		"IPv6 NS address fetches":      "GlueFetchv6",
		"queries with RTT 10-100ms":    "QryRTTnn",
		"queries with RTT 100-500ms":   "QryRTTnn",
		"queries with RTT 500-800ms":   "QryRTTnn",
		"queries with RTT 800-1600ms":  "QryRTTnn",
		"queries with RTT < 10ms":      "QryRTTnn",
		"queries with RTT > 1600ms":    "QryRTTnn",
		"query retries":                "Retry",
		"query timeouts":               "QueryTimeout",
		"truncated responses received": "Truncated",
	}
	socketMap = map[string]string{
		"Raw sockets opened":               "Raw_Open",
		"Raw sockets active":               "Raw_Active",
		"TCP IPv4 connections accepted":    "TCPv4_Accept",
		"TCP IPv4 sockets active":          "TCPv4_Active",
		"TCP IPv4 sockets closed":          "TCPv4_Close",
		"TCP IPv4 sockets opened":          "TCPv4_Open",
		"TCP IPv6 socket bind failures":    "TCPv6_BindFail",
		"TCP IPv6 sockets closed":          "TCPv6_Close",
		"TCP IPv6 sockets opened":          "TCPv6_Open",
		"UDP IPv4 connections established": "UDPv4_Conn",
		"UDP IPv4 send errors":             "UDPv4_SendErr",
		"UDP IPv4 sockets active":          "UDPv4_Active",
		"UDP IPv4 sockets closed":          "UDPv4_Close",
		"UDP IPv4 sockets opened":          "UDPv4_Open",
	}
	zoneMap = map[string]string{
		"IPv6 notifies sent":               "NotifyOutv6",
		"IPv6 notifies received":           "NotifyInv6",
		"IPv6 SOA queries sent":            "SOAOutv6",
		"IPv6 AXFR requested":              "AXFRReqv6",
		"IPv6 IXFR requested":              "IXFRReqv6",
		"IPv4 IXFR requested":              "IXFRReqv4",
		"IPv4 SOA queries sent":            "SOAOutv4",
		"IPv4 notifies received":           "NotifyInv4",
		"IPv4 notifies sent":               "NotifyOutv4",
		"IPv4 AXFR requested":              "AXFRReqv4",
		"notifies rejected":                "NotifyRej",
		"Incoming notifies rejected":       "NotifyRej",
		"transfer requests succeeded":      "XfrSuccess",
		"Zone transfer requests succeeded": "XfrSuccess",
		"Zone transfer requests failed":    "XfrFail",
		"transfer requests failed":         "XfrFail",
	}
	cacheStatsMap = map[string]string{
		"cache database hash buckets":                    "Buckets",
		"cache database nodes":                           "Nodes",
		"cache heap highest memory in use":               "UseHeapHighest",
		"cache heap memory in use":                       "UseHeapMemory",
		"cache heap memory total":                        "TotalHeapMemory",
		"cache hits":                                     "Hits",
		"cache hits (from query)":                        "QueryHits",
		"cache misses":                                   "Misses",
		"cache misses (from query)":                      "QueryMisses",
		"cache records deleted due to TTL expiration":    "DelTTL",
		"cache records deleted due to memory exhaustion": "DelMem",
		"cache tree highest memory in use":               "UseTreeHighest",
		"cache tree memory in use":                       "UseTreeMemory",
		"cache tree memory total":                        "TotalTreeMemory",
	}
)

type statsCollector struct {
	filePath string
	rndc     string
}

// newServerCollector implements collectorConstructor.
func NewStatsCollector(fd, rndc string) prometheus.Collector {
	return &statsCollector{
		filePath: fd,
		rndc:     rndc,
	}
}

// Describe implements prometheus.Collector.
func (c *statsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- bootTime
	ch <- nameServerStatistics
	ch <- incomingQueries
	ch <- incomingRequests
	ch <- incomingRequests
	ch <- socketIO
	ch <- zoneMetricStats
	ch <- cacheRRsetsStats
	ch <- cacheStatistics

	for _, desc := range resolverMetricStatsFile {
		ch <- desc
	}
	for _, desc := range cacheMetricStatsFile {
		ch <- desc
	}
}

// Collect implements prometheus.Collector.
func (c *statsCollector) Collect(ch chan<- prometheus.Metric) {
	var outInfo bytes.Buffer
	rcmd := exec.Command("/bin/sh", c.rndc)
	rcmd.Stdout = &outInfo
	err := rcmd.Run()
	log.Info("sh info:", outInfo.String())
	if err != nil {
		log.Error(err)
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		return
	}
	contentBs, err := ioutil.ReadFile(c.filePath)
	if err != nil || len(contentBs) < 10 {
		log.Error(err)
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		return
	}
	statsInfo := ParserStats(string(contentBs))
	// log.Println(statsInfo)
	ch <- prometheus.MustNewConstMetric(
		bootTime, prometheus.GaugeValue, float64(statsInfo.BootTime),
	)
	if mds, ok := statsInfo.ModuleMap["Incoming Requests"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				ch <- prometheus.MustNewConstMetric(
					incomingRequests, prometheus.GaugeValue, value, key,
				)
			}
		}
	}
	if mds, ok := statsInfo.ModuleMap["Incoming Queries"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				ch <- prometheus.MustNewConstMetric(
					incomingQueries, prometheus.CounterValue, value, key,
				)
			}
		}
	}
	// Name Server Statistics
	if mds, ok := statsInfo.ModuleMap["Name Server Statistics"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				if tk, kok := nameServerMap[key]; kok {
					ch <- prometheus.MustNewConstMetric(
						nameServerStatistics, prometheus.CounterValue, value, tk,
					)
				}
			}
		}
	}
	// Outgoing Queries
	if mds, ok := statsInfo.ModuleMap["Outgoing Queries"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				ch <- prometheus.MustNewConstMetric(
					outgoingQueries, prometheus.CounterValue, value, md.View, key,
				)
			}
		}
	}
	// Resolver Statistics  resolverQueries
	if mds, ok := statsInfo.ModuleMap["Resolver Statistics"]; ok {
		for _, md := range mds {

			for key, value := range md.Info {
				if rk, kok := resolverStatisticsMap[key]; kok {
					if rk != "QryRTTnn" {
						if pd, pok := resolverMetricStatsFile[rk]; pok {
							ch <- prometheus.MustNewConstMetric(
								pd, prometheus.CounterValue, value, md.View,
							)
						}
					}
				}
			}
			if pd, pok := resolverMetricStatsFile["QryRTTnn"]; pok {
				if buckets, count, err := getHistogram(md); err == nil {
					ch <- prometheus.MustNewConstHistogram(
						pd, count, math.NaN(), buckets, md.View,
					)
				}
			}
		}
	}

	// Socket IO Statistics
	if mds, ok := statsInfo.ModuleMap["Socket IO Statistics"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				if tk, kok := socketMap[key]; kok {
					ch <- prometheus.MustNewConstMetric(
						socketIO, prometheus.CounterValue, value, tk,
					)
				}
			}
		}
	}
	// Zone Maintenance Statistics
	if mds, ok := statsInfo.ModuleMap["Zone Maintenance Statistics"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				if tk, kok := zoneMap[key]; kok {
					ch <- prometheus.MustNewConstMetric(
						zoneMetricStats, prometheus.CounterValue, value, tk,
					)
				}
			}
		}
	}
	// Cache DB RRsets
	if mds, ok := statsInfo.ModuleMap["Cache DB RRsets"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				/*if len(md.View) < 1 {
					continue
				}*/
				idx := strings.Index(md.View, "(")
				if idx < 0 {
					idx = len(md.View)
				}
				view := strings.Trim(md.View[0:idx], " ")
				ch <- prometheus.MustNewConstMetric(
					cacheRRsetsStats, prometheus.CounterValue, value, view, key,
				)
			}
		}
	}
	// Cache Statistics
	if mds, ok := statsInfo.ModuleMap["Cache Statistics"]; ok {
		for _, md := range mds {
			for key, value := range md.Info {
				if rk, kok := cacheStatsMap[key]; kok {
					if pd, pok := cacheMetricStatsFile[rk]; pok {
						ch <- prometheus.MustNewConstMetric(
							pd, prometheus.GaugeValue, value, md.View,
						)
					}
				}
			}
		}
	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)
}

type RttHistog struct {
	key string
	le  float64
}

var (
	rttList = []RttHistog{
		RttHistog{"queries with RTT < 10ms", 10},
		RttHistog{"queries with RTT 10-100ms", 100},
		RttHistog{"queries with RTT 100-500ms", 500},
		RttHistog{"queries with RTT 500-800ms", 800},
		RttHistog{"queries with RTT 800-1600ms", 1600},
		RttHistog{"queries with RTT > 1600ms", 2000},
	}
)

func getHistogram(md Module) (map[float64]uint64, uint64, error) {
	buckets := map[float64]uint64{}
	var count uint64
	for _, rtt := range rttList {
		if value, ok := md.Info[rtt.key]; ok {
			buckets[rtt.le] = count + uint64(value)
			count += uint64(value)
		}
	}
	return buckets, count, nil
}

// easygen: json
type Module struct {
	View string             `json:"view"`
	Info map[string]float64 `json:"info"`
}

// easygen: json
type StatusInfo struct {
	BootTime  int64               `json:"boot_time"`
	ModuleMap map[string][]Module `json:"module_map"`
}

func ParserStats(content string) *StatusInfo {

	lines := strings.Split(content, "\n")
	sub := ""
	stats := StatusInfo{
		ModuleMap: map[string][]Module{},
	} // map[string][]Module{}
	ts := []string{}
	im := &Module{
		Info: map[string]float64{},
	}
	view := ""
	for _, line := range lines {
		num := []string{}
		zimu := []string{}
		if strings.HasPrefix(line, "+++") {
			// 提取时间戳
			ts = numReg.FindAllString(line, -1)
		} else if strings.HasPrefix(line, "---") {
			break
		} else if strings.HasPrefix(line, "++") {
			// sub = ""
			if len(im.Info) > 0 || len(im.View) > 0 {
				stats.ModuleMap[sub] = append(stats.ModuleMap[sub], *im)
				im = &Module{
					Info: map[string]float64{},
					View: "",
				}
			}
			im.View = ""
			sub = subReg.ReplaceAllString(line, "")
			sub = viewReg.ReplaceAllString(sub, "")
			// sub = sss3.ReplaceAllString(sub, "")

		} else if strings.HasPrefix(line, "[") {
			view = strings.ReplaceAll(line, "[", "")
			view = strings.ReplaceAll(view, "View:", "")
			view = strings.ReplaceAll(view, "]", "")
			view = strings.Trim(view, " ")
			if len(im.Info) > 0 {
				stats.ModuleMap[sub] = append(stats.ModuleMap[sub], *im)
			}
			im = &Module{
				Info: map[string]float64{},
				View: view,
			}
			// fmt.Println(sub, view)
		} else {
			num = numReg.FindAllString(line, 1)
			/*zimu = letReg.FindAllString(line, -1)
			if len(num) > 0 && len(zimu) > 0 {
				v, _ := strconv.ParseFloat(num[0], 10)
				im.Info[strings.Join(zimu, " ")] = v
			}*/
			if len(num) > 0 {
				line = strings.Replace(line, num[0], "", 1)
				zimu = letReg.FindAllString(line, -1)
				if len(zimu) > 0 {
					v, _ := strconv.ParseFloat(num[0], 10)
					im.Info[strings.Join(zimu, " ")] = v
				}
			}
		}
	}
	if len(ts) > 0 {
		ti, _ := strconv.ParseInt(ts[0], 10, 64)
		stats.BootTime = ti
	}

	/*fmt.Println(stats, ts[0])
	bs, _ := json.Marshal(stats)
	fmt.Println(string(bs))*/
	return &stats
}
