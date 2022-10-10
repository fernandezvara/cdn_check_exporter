# cdn_check_exporter

CDN check is an Prometheus exporter that helps monitoring assets served by a content delivery network.

When serving content using a CDN, many hosts will serve the content, so there is no (normally) a fixed set of IPs to query to. This service queries DNS to discover IPs and then queries the hosts for the content. 

Basically it makes `curl http://123.123.123.13/resource -H "Host: www.example.com"` but in an automated way that allows to find if the content is the expected (by looking at the bytes returned), the error code returned and its availabilty. 

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
DOMAIN="www.google.com" RESOURCE="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png" ./cdn_check_exporter

```

Output:

``` 
curl http://127.0.0.1:9000/metrics

# HELP dd_request request error code
# TYPE dd_request gauge
dd_request{ip="142.250.184.4",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 200
dd_request{ip="142.250.201.68",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 200
# HELP dd_request_bytes request length
# TYPE dd_request_bytes gauge
dd_request_bytes{ip="142.250.184.4",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 13504
dd_request_bytes{ip="142.250.201.68",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 13504
# HELP dd_request_last_update request last update for this IP
# TYPE dd_request_last_update gauge
dd_request_last_update{ip="142.250.184.4",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 1.665413412e+09
dd_request_last_update{ip="142.250.201.68",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 1.665413412e+09
# HELP dd_request_success request was success
# TYPE dd_request_success gauge
dd_request_success{ip="142.250.184.4",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 1
dd_request_success{ip="142.250.201.68",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 1
# HELP dd_request_time request time in ms
# TYPE dd_request_time gauge
dd_request_time{ip="142.250.184.4",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 25
dd_request_time{ip="142.250.201.68",resource="/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png"} 27
...
```


