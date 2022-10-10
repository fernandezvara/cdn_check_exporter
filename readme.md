# cdn_check_exporter

CDN check is an exporter that helps monitoring assets served by a content delivery network.

Every **x seconds** this service will query to the discovered IPs serving traffic to the domain for the requested resources. 


```
go run main.go -h
Usage of /tmp/go-build3023912900/b001/exe/main:
  -debug=false: see that the service is doing
  -domain="": domain to query to
  -every=10: number of seconds to wait between requests
  -metrics.ip="0.0.0.0:9000": IP and Port for the metrics exporter
  -metrics.path="/metrics": path under metrics are exposed
  -resource="": resource(s) to query to, separated by commas
```

Every flag can be passed as *environment variable*, sample:

```
DEBUG=1 DOMAIN="www.google.com" RESOURCE="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png" ./cdn_check_exporter
