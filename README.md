# bind_stats_exporter
Get BIND DNS stats from named_stats file and prometheus.

# Why do we need bind_stats_exporter?
We all know that [bind_exporter](https://github.com/prometheus-community/bind_exporter)  , in use we find that the `Bind DNS` is collected through the interface, whether it is using `XML` or `json`. It will cause the `Bind DNS` to be stuck. In other words, the collection indicator affects the availability of `Bind DNS`.
After our verification, we found that using `statistics-file` and `rndc`, there is no blocking problem with `Bind DNS`. This is the point of existence of `bind_stats_exporter`.

# How to use
## Build and run from source
```shell script
go get github.com/qiangmzsx/bind_stats_exporter
cd $GOPATH/src/github.com/qiangmzsx/bind_stats_exporter
go build -v
./bind_stats_exporter [flags]
```

## Configure BIND
``` 
options {
  ... ...
  statistics-file "/var/named/named.stats";
  ... ...
}
```

## stats.
The purpose of `stats.sh` is to trigger Bind DNS to output statistics to the specified file, although some other operations can also be performed, 
but mainly to trigger the generation of statistics files.

```shell script
echo "" > /var/named/named.stats
echo "echo OK"
/usr/sbin/rndc stats
echo "rndc OK"
```

## start
```shell script
./bind_stats_exporter --bind.stats-file=/var/named/named.stats  --bind.sh=./stats.sh 
```
```shell script
curl http://ip:9219/metrics

```

