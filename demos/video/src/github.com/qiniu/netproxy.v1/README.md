github.com/qiniu/netproxy.v1
===================

# 概述

要解决的需求：

* 服务发现。
* 服务自动迁移。

# 使用

## qiniunetproxy

TODO

## 配置文件 (json)

```json
{
  "proxys": [
    {
      "service": "rs.qiniuapi.com",
      "as": "http://localhost",
      "host": "rs.qiniuapi.com"
    },
    {
      "service": "rs.master.mongo",
      "as": "tcp://localhost:27017"
    },
    {
      "service": "udp.example",
      "as": "udp://localhost:1234"
    }
  ]
}
```
