# tRPC-Go prometheus metrics plugin 

English | [中文](README_CN.md)

## Config
```yaml
plugins:                                          #Plugin configuration.
  metrics:                                        #Reference metrics.
    prometheus:                                   #Start prometheus.
      ip: 0.0.0.0                                 #Prometheus bind address.
      port: 8090                                  #Prometheus bind port.
      path: /metrics                              #Metrics path.
      namespace: Development                      #Namespace.
      subsystem: trpc                             #Subsystem.
      rawmode:   false                            #Raw mode, no conversion of special characters for metrics.
      enablepush: true                            #Enable push mode, not enabled by default.
      gateway: http://localhost:9091              #Prometheus gateway address.
      password: username:MyPassword               #Set the account password, username and MyPassword are split by a colon.
      job: job                                    #Job name.
      pushinterval: 1                             #Push data every 1 second by default
```

## Tutorial
### Reference prometheus
Reference it in main.go and configure the parameters in yaml.

```golang
import _ "trpc.group/trpc-go/trpc-metrics-prometheus"
```

### Report data
trpc metrics usage guidelines [trpc metrics](https://github.com/trpc-group/trpc-go/blob/main/metrics/README.md)

## Query reported data
Query the metrics locally via curl to see if the metrics were generated successfully.
```bash
curl ${ip}:${port}/$path |grep ${namespace}_${subsystem}
```

## Report call data
Add configuration
```yaml
  filter:
    - prometheus                                   #Add prometheus filter
```
The call data currently supports both Histogram for request time and SUM for request volume.
The metric names are prefixed with ClientFilter and ServerFilter.


## Notice
1. Prometheus currently only supports PolicySUM/PolicySET/PolicyHistogram type reporting, other types of support please submit pr.
2. Prometheus metric does not support Chinese and special characters, illegal characters will be automatically converted to '_' in the acsii table, Chinese and other utf8 characters are converted to the corresponding data, such as "trpc.Chinese metric" -> "trpc_20013_25991_25351_26631_", close this function can be used to set rawmode is true, exception reporting will fail directly.
3. The plugin only provides exporter, not Pushgateway and Prometheus server.
4. Multi-dimension reporting uses the metrics.NewMultiDimensionMetricsX interface to set multi-dimension names, otherwise conflicts may occur.
5. If you need to push custom data, you can call the GetDefaultPusher method after the plugin is initialized, otherwise the returned pusher is empty.
