# iplookup

This filter implements a generic ip lookup method using the package [geoiplookup](https://bitbucket.org/HelgeOlav/geoiplookup).
All lookups are done using gRPC to another server (either locally or remote) that has implemented any geoiplookup server.

```json
{
  "filter": [
    {
      "type": "iplookup",
      "server": "localhost:8081",
      "timeout": 2000,
      "ip_field": "message",
      "key": "iplocation",
      "skip_private": true,
      "private_net": [
        "172.16.0.0/12"
      ]
    }
  ]
}
```

Private nets are skipped by default, and all private IPv4 and IPv6 are included in the default list. Timeout is the time to wait for an answer in milliseconds, default is 2 sec.