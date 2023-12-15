package main

import (
	"encoding/json"
	"strings"
	"testing"
)

var (
	str = `
+++ Statistics Dump +++ (1598003941)
++ Incoming Requests ++
          2263459024 QUERY
               31342 NOTIFY
++ Incoming Queries ++
          1188374276 A
               31663 NS
                1307 SOA
            42634454 PTR
                  75 HINFO
             2628346 MX
                  23 TXT
          1029784471 AAAA
                4177 SRV
                   5 TYPE65
                 152 ANY
                  61 Others
++ Outgoing Queries ++
[View: view_bj_ali]
           102201078 A
            19937398 NS
              376663 PTR
            25492975 AAAA
                3189 SRV
                   1 Others
[View: view_bj_ali_test]
            33445365 A
             5309953 NS
                3996 PTR
             7283514 AAAA
                   5 SRV
[View: view_bj_mjq]
            80762809 A
            14270909 NS
              330987 PTR
                  39 HINFO
               51535 MX
            21276772 AAAA
                  75 SRV
                  93 ANY
                 951 Others
[View: view_bj_mjq_test]
            11322547 A
             2758143 NS
                4605 PTR
                  36 HINFO
             4365341 AAAA
                   6 SRV
[View: any]
             1500221 A
              705192 NS
                 419 SOA
                2336 PTR
               94650 MX
                  13 TXT
              242118 AAAA
                  20 SRV
                   5 TYPE65
[View: _bind]
++ Name Server Statistics ++
          2263490366 IPv4 requests received
               28268 requests with EDNS(0) received
               19221 requests with TSIG received
             2349966 TCP requests received
          2262987116 responses sent
             2328180 truncated responses sent
               28265 responses with EDNS(0) sent
          1251364043 queries resulted in successful answer
           202428284 queries resulted in authoritative answer
          2060239985 queries resulted in non authoritative answer
                 169 queries resulted in referral answer
           830159161 queries resulted in nxrrset
              287491 queries resulted in SERVFAIL
           181144896 queries resulted in NXDOMAIN
           280191101 queries caused recursion
              216291 duplicate queries received
              286959 queries dropped
++ Zone Maintenance Statistics ++
               30300 IPv4 notifies sent
               24667 IPv4 notifies received
               18038 notifies rejected
              100406 IPv4 SOA queries sent
                6275 IPv4 IXFR requested
                6275 transfer requests succeeded
++ Resolver Statistics ++
[Common]
                   1 mismatch responses received
[View: view_bj_ali]
           148011304 IPv4 queries sent
           147134252 IPv4 responses received
              391209 NXDOMAIN received
               31273 SERVFAIL received
                1334 EDNS(0) query failures
               54698 truncated responses received
              909991 query retries
              877052 query timeouts
                  75 IPv4 NS address fetches
                   1 IPv6 NS address fetches
           134115772 queries with RTT < 10ms
            12813170 queries with RTT 10-100ms
              151823 queries with RTT 100-500ms
               27748 queries with RTT 500-800ms
               25058 queries with RTT 800-1600ms
                 681 queries with RTT > 1600ms
[View: view_bj_ali_test]
            46042833 IPv4 queries sent
            45956474 IPv4 responses received
              147984 NXDOMAIN received
                 446 SERVFAIL received
                   9 EDNS(0) query failures
               14100 truncated responses received
              100579 query retries
               86359 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
            41692546 queries with RTT < 10ms
             4134354 queries with RTT 10-100ms
              106909 queries with RTT 100-500ms
               16378 queries with RTT 500-800ms
                6286 queries with RTT 800-1600ms
                   1 queries with RTT > 1600ms
[View: view_bj_mjq]
           116694170 IPv4 queries sent
           115627003 IPv4 responses received
              396072 NXDOMAIN received
              214824 SERVFAIL received
                 834 other errors received
                4935 EDNS(0) query failures
               40866 truncated responses received
             1144313 query retries
             1067167 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
           102118865 queries with RTT < 10ms
            13205535 queries with RTT 10-100ms
              230305 queries with RTT 100-500ms
               29789 queries with RTT 500-800ms
               40530 queries with RTT 800-1600ms
                1979 queries with RTT > 1600ms
[View: view_bj_mjq_test]
            18450678 IPv4 queries sent
            18420614 IPv4 responses received
               13507 NXDOMAIN received
                1412 SERVFAIL received
                  13 EDNS(0) query failures
                4773 truncated responses received
               36040 query retries
               30064 query timeouts
                   3 IPv4 NS address fetches
                   1 IPv6 NS address fetches
            17031735 queries with RTT < 10ms
             1330444 queries with RTT 10-100ms
               45991 queries with RTT 100-500ms
                6523 queries with RTT 500-800ms
                5920 queries with RTT 800-1600ms
                   1 queries with RTT > 1600ms
[View: any]
             2544974 IPv4 queries sent
             2508794 IPv4 responses received
                5718 NXDOMAIN received
                5696 SERVFAIL received
                  50 EDNS(0) query failures
                2022 truncated responses received
               38204 query retries
               36180 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
             2188811 queries with RTT < 10ms
              304379 queries with RTT 10-100ms
               12852 queries with RTT 100-500ms
                 591 queries with RTT 500-800ms
                2146 queries with RTT 800-1600ms
                  15 queries with RTT > 1600ms
[View: _bind]
++ Cache DB RRsets ++
[View: view_bj_ali (Cache: view_bj_ali)]
                1031 A
                   1 NS
                  64 CNAME
                   3 AAAA
                   1 RRSIG
                1020 !AAAA
                  18 NXDOMAIN
[View: view_bj_ali_test (Cache: view_bj_ali_test)]
                 577 A
                   1 NS
                  33 CNAME
                   2 AAAA
                   1 RRSIG
                 572 !AAAA
                   9 NXDOMAIN
[View: view_bj_mjq (Cache: view_bj_mjq)]
                 918 A
                   1 NS
                  48 CNAME
                   1 PTR
                   3 AAAA
                   1 RRSIG
                   1 !MX
                 962 !AAAA
                  73 NXDOMAIN
[View: view_bj_mjq_test (Cache: view_bj_mjq_test)]
                 450 A
                   1 NS
                  36 CNAME
                   1 AAAA
                   1 RRSIG
                 494 !AAAA
                   4 NXDOMAIN
[View: any (Cache: any)]
                  10 A
                   1 NS
                  14 CNAME
                   1 AAAA
                   1 RRSIG
                   2 !MX
                   1 !AAAA
                   2 NXDOMAIN
[View: default]
70 A
11 NS
2 SOA
64 AAAA
5 DS
14 RRSIG
1 NSEC
2 DNSKEY
1 !NS
2 !AAAA
1 !DS
[View: ]
70 A
11 NS
2 SOA
64 AAAA
5 DS
14 RRSIG
1 NSEC
2 DNSKEY
1 !NS
2 !AAAA
1 !DS
[View: _bind (Cache: _bind)]
++ Socket I/O Statistics ++
           331732008 UDP/IPv4 sockets opened
              122738 TCP/IPv4 sockets opened
                   1 Raw sockets opened
           331731784 UDP/IPv4 sockets closed
             2416816 TCP/IPv4 sockets closed
                1496 UDP/IPv4 socket bind failures
           331627505 UDP/IPv4 connections established
              122725 TCP/IPv4 connections established
             2294088 TCP/IPv4 connections accepted
                 108 UDP/IPv4 send errors
                   1 TCP/IPv4 send errors
++ Per Zone Query Statistics ++
--- Statistics Dump --- (1598003941)

`
	str2 = `
+++ Statistics Dump +++ (1598326557)
++ Incoming Requests ++
++ Incoming Queries ++
++ Outgoing Rcodes ++
++ Outgoing Queries ++
[View: view_bj_ali]
                   1 A
                   2 AAAA
[View: view_bj_ali_test]
                   2 A
                   2 AAAA
[View: view_bj_mjq]
                   1 A
                   1 AAAA
[View: view_bj_mjq_test]
                   1 A
                   2 AAAA
[View: any]
                   2 A
                   2 AAAA
[View: _bind]
++ Name Server Statistics ++
++ Zone Maintenance Statistics ++
                  25 IPv4 notifies sent
++ Resolver Statistics ++
[Common]
[View: view_bj_ali]
                   3 IPv4 queries sent
                   2 IPv4 responses received
                   1 query retries
                   1 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
                   2 queries with RTT 10-100ms
                 523 bucket size
[View: view_bj_ali_test]
                   4 IPv4 queries sent
                   2 IPv4 responses received
                   2 query retries
                   2 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
                   2 queries with RTT 10-100ms
                 523 bucket size
[View: view_bj_mjq]
                   2 IPv4 queries sent
                   2 IPv4 responses received
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
                   2 queries with RTT 10-100ms
                 523 bucket size
[View: view_bj_mjq_test]
                   3 IPv4 queries sent
                   2 IPv4 responses received
                   1 query retries
                   1 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
                   1 queries with RTT 10-100ms
                   1 queries with RTT 100-500ms
                 523 bucket size
[View: any]
                   4 IPv4 queries sent
                   2 IPv4 responses received
                   2 query retries
                   2 query timeouts
                   1 IPv4 NS address fetches
                   1 IPv6 NS address fetches
                   2 queries with RTT 10-100ms
                 523 bucket size
[View: _bind]
                 523 bucket size
++ Cache Statistics ++
[View: view_bj_ali (Cache: view_bj_ali)]
                   0 cache hits
                   2 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   1 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29840 cache tree memory in use
               30096 cache tree highest memory in use
              270336 cache heap memory total
                9216 cache heap memory in use
                9216 cache heap highest memory in use
[View: view_bj_ali_test (Cache: view_bj_ali_test)]
                   0 cache hits
                   2 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   1 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29848 cache tree memory in use
               30104 cache tree highest memory in use
              270336 cache heap memory total
                9216 cache heap memory in use
                9216 cache heap highest memory in use
[View: view_bj_mjq (Cache: view_bj_mjq)]
                   0 cache hits
                   2 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   1 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29840 cache tree memory in use
               30096 cache tree highest memory in use
              270336 cache heap memory total
                9216 cache heap memory in use
                9216 cache heap highest memory in use
[View: view_bj_mjq_test (Cache: view_bj_mjq_test)]
                   0 cache hits
                   2 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   1 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29848 cache tree memory in use
               30104 cache tree highest memory in use
              270336 cache heap memory total
                9216 cache heap memory in use
                9216 cache heap highest memory in use
[View: any (Cache: any)]
                   0 cache hits
                   2 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   1 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29832 cache tree memory in use
               30088 cache tree highest memory in use
              270336 cache heap memory total
                9216 cache heap memory in use
                9216 cache heap highest memory in use
[View: _bind (Cache: _bind)]
                   0 cache hits
                   0 cache misses
                   0 cache hits (from query)
                   0 cache misses (from query)
                   0 cache records deleted due to memory exhaustion
                   0 cache records deleted due to TTL expiration
                   0 cache database nodes
                  64 cache database hash buckets
              287408 cache tree memory total
               29512 cache tree memory in use
               29512 cache tree highest memory in use
              262144 cache heap memory total
                1024 cache heap memory in use
                1024 cache heap highest memory in use
++ Cache DB RRsets ++
[View: view_bj_ali (Cache: view_bj_ali)]
                   1 CNAME
[View: view_bj_ali_test (Cache: view_bj_ali_test)]
                   1 CNAME
[View: view_bj_mjq (Cache: view_bj_mjq)]
                   1 CNAME
[View: view_bj_mjq_test (Cache: view_bj_mjq_test)]
                   1 CNAME
[View: any (Cache: any)]
                   1 CNAME
[View: _bind (Cache: _bind)]
++ ADB stats ++
[View: view_bj_ali]
                1021 Address hash table size
                   2 Addresses in hash table
                1021 Name hash table size
                   2 Names in hash table
[View: view_bj_ali_test]
                1021 Address hash table size
                   2 Addresses in hash table
                1021 Name hash table size
                   2 Names in hash table
[View: view_bj_mjq]
                1021 Address hash table size
                   2 Addresses in hash table
                1021 Name hash table size
                   2 Names in hash table
[View: view_bj_mjq_test]
                1021 Address hash table size
                   2 Addresses in hash table
                1021 Name hash table size
                   2 Names in hash table
[View: any]
                1021 Address hash table size
                   2 Addresses in hash table
                1021 Name hash table size
                   2 Names in hash table
[View: _bind]
                1021 Address hash table size
                1021 Name hash table size
++ Socket I/O Statistics ++
                  25 UDP/IPv4 sockets opened
                   7 TCP/IPv4 sockets opened
                   1 TCP/IPv6 sockets opened
                   1 Raw sockets opened
                  20 UDP/IPv4 sockets closed
                   7 TCP/IPv4 sockets closed
                   1 TCP/IPv6 sockets closed
                   1 TCP/IPv6 socket bind failures
                  16 UDP/IPv4 connections established
                   8 TCP/IPv4 connections accepted
                   2 UDP/IPv4 send errors
                   5 UDP/IPv4 sockets active
                   8 TCP/IPv4 sockets active
                   1 Raw sockets active
++ Per Zone Query Statistics ++
--- Statistics Dump --- (1598326557)
`
)

func Test_ParserStats(t *testing.T) {

	si := ParserStats(str)
	bs, _ := json.Marshal(si)
	t.Log(string(bs))
}

func Test_ParserStatsForRRsets(t *testing.T) {

	statsInfo := ParserStats(str)
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
				t.Log(value, view, key)
			}
		}
	}
}
