[English](README.md) | 中文

# tRPC-Go prometheus metrics plugin 

[![Go Reference](https://pkg.go.dev/badge/github.com/trpc-ecosystem/go-metrics-prometheus.svg)](https://pkg.go.dev/github.com/trpc-ecosystem/go-metrics-prometheus)
[![Go Report Card](https://goreportcard.com/badge/github.com/trpc.group/trpc-go/trpc-metrics-prometheus)](https://goreportcard.com/report/github.com/trpc.group/trpc-go/trpc-metrics-prometheus)
[![LICENSE](https://img.shields.io/github/license/trpc-ecosystem/go-metrics-prometheus.svg?style=flat-square)](https://github.com/trpc-ecosystem/go-metrics-prometheus/blob/main/LICENSE)
[![Releases](https://img.shields.io/github/release/trpc-ecosystem/go-metrics-prometheus.svg?style=flat-square)](https://github.com/trpc-ecosystem/go-metrics-prometheus/releases)
[![Docs](https://img.shields.io/badge/docs-latest-green)](http://test.trpc.group.woa.com/docs/)
[![Tests](https://github.com/trpc-ecosystem/go-metrics-prometheus/actions/workflows/prc.yaml/badge.svg)](https://github.com/trpc-ecosystem/go-metrics-prometheus/actions/workflows/prc.yaml)
[![Coverage](https://codecov.io/gh/trpc-ecosystem/go-metrics-prometheus/branch/main/graph/badge.svg)](https://app.codecov.io/gh/trpc-ecosystem/go-metrics-prometheus/tree/main)

## 配置
```yaml
plugins:                                          #插件配置
  metrics:                                        #引用metrics
    prometheus:                                   #启动prometheus
      ip: 0.0.0.0                                 #prometheus绑定地址
      port: 8090                                  #prometheus绑定端口
      path: /metrics                              #metrics路径
      namespace: Development                      #命名空间
      subsystem: trpc                             #子系统
      rawmode:   false                            #原始模式，不会对metrics的特殊字符进行转换 
      enablepush: true                            #启用push模式，默认不启用
      gateway: http://localhost:9091              #prometheus gateway地址
      password: username:MyPassword               #设置账号密码， 以冒号分割
      job: job                                    #job名称
      pushinterval: 1                             #push间隔，默认1s上报一次
```

## 教程
### 引入prometheus
在main.go中引用,并在yaml中配置好参数

```golang
import _ "trpc.group/trpc-go/trpc-metrics-prometheus"
```

### 上报数据
trpc metrics 使用指引 [trpc metrics](https://github.com/trpc-group/trpc-go/blob/main/metrics/README_CN.md)

## 查询上报数据
本地通过curl查询指标，查看指标是否生成成功
```bash
curl ${ip}:${port}/$path |grep ${namespace}_${subsystem}
```

## 上报调用数据
增加配置
```yaml
  filter:
    - prometheus                                   #增加prometheus filter
```
调用数据目前支持请求耗时的Histogram和请求量的SUM两个指标
指标名前缀为ClientFilter与ServerFilter


## 注意事项
1. prometheus目前只支持PolicySUM/PolicySET/PolicyHistogram类型上报，其它类型支持请提pr
2. prometheus指标不支持中文与特殊字符，非法字符将会自动转换，acsii表内的非法字符转换为'_'，中文等utf8字符转换为对应的数据，比如"trpc.中文指标"->"trpc_20013_25991_25351_26631_",关闭此功能可以使用设置rawmode为true，异常上报将直接失败
3. 插件只提供exporter，不提供平台与对接
4. 多维度上报使用 metrics.NewMultiDimensionMetricsX 接口设置多维度名，否则可能会出现冲突
5. 如果需要推送自定义数据，可以在插件初始化完之后调用GetDefaultPusher方法，否则返回的pusher为空
