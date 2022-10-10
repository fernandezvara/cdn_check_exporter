package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// exposed holds the various metrics that are collected
	exposed = map[string]*prometheus.GaugeVec{}
	// show last update time to see if system is working correctly
	lastUpdate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dd_lastUpdate",
		Help: "Last update timestamp in epoch seconds",
	},
		[]string{"scope"},
	)
	domain      string
	resource    string
	resources   []string
	every       int
	metricsIP   string
	metricsPath string
	// metricsCert string
	// metricsPriv string
	debugFlag bool
	ips       []string
)

func addString(item string, items []string) (newItems []string) {

	newItems = items

	for _, i := range items {
		if i == item {
			return
		}
	}

	newItems = append(newItems, item)
	return

}

func main() {

	flag.StringVar(&domain, "domain", "", "domain to query to")
	flag.StringVar(&resource, "resource", "", "resource(s) to query to, separated by commas")
	flag.IntVar(&every, "every", 10, "number of seconds to wait between requests")
	flag.StringVar(&metricsIP, "ipport", "0.0.0.0:9000", "IP and Port for the metrics exporter")
	flag.StringVar(&metricsPath, "path", "/metrics", "path under metrics are exposed")
	// flag.StringVar(&metricsCert, "metrics.cert", "", "certificate used to expose metrics")
	// flag.StringVar(&metricsPriv, "metrics.priv", "", "private key used to expose metrics")
	flag.BoolVar(&debugFlag, "debug", false, "see that the service is doing")

	flag.Parse()

	// parse resource -> resources
	resources = strings.Split(resource, ",")

	if len(domain) == 0 || len(resources) == 0 {
		fmt.Println("Missing required values of domain and/or resources")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// install promhttp handler for metricsPath (/metrics)
	http.Handle(metricsPath, promhttp.Handler())

	// show nice web page if called without metricsPath
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
					<head><title>CDN check Exporter</title></head>
					<body>
					<h1>CDN check Exporter</h1>
					<p><a href='` + metricsPath + `'>Metrics</a></p>
					</body>
					</html>`))
	})

	// Start the http server in background, but catch error
	go func() {
		err := http.ListenAndServe(metricsIP, nil)
		fmt.Println("err:", err)
		os.Exit(2)
	}()

	// wait for initialization of http server before looping so the systemd alive check doesn't fail
	time.Sleep(time.Second * 3)

	// notify systemd that we're ready
	daemon.SdNotify(false, daemon.SdNotifyReady)

	for {
		work(domain, resources)

		systemAlive(metricsIP, metricsPath)

		time.Sleep(time.Duration(int64(every)) * time.Second)
	}

}

func work(domain string, resources []string) {

	requests, err := net.LookupIP(domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range requests {
		ips = addString(ip.String(), ips)
	}

	printLog(len(ips), "IPs found ---")

	for _, ip := range ips {

		for _, resource := range resources {

			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s%s", ip, resource), nil)
			req.Host = domain

			var start time.Time = time.Now()

			res, err := client.Do(req)

			labelKeys := []string{"ip", "resource"}

			var labelValues []string
			labelValues = append(labelValues, string(ip))
			labelValues = append(labelValues, resource)

			if err != nil {
				printLog(ip, err)
				setPrometheusMetric("request_success", 0, labelKeys, labelValues, "request was success")

			} else {
				printLog(ip, res.StatusCode)
				body, _ := io.ReadAll(res.Body)
				printLog(ip, string(body))
				res.Body.Close()

				setPrometheusMetric("request_success", 1, labelKeys, labelValues, "request was success")
				setPrometheusMetric("request", res.StatusCode, labelKeys, labelValues, "request error code")
				setPrometheusMetric("request_time", int(time.Since(start).Milliseconds()), labelKeys, labelValues, "request time in ms")
				setPrometheusMetric("request_bytes", len(body), labelKeys, labelValues, "request length")
				setPrometheusMetric("request_last_update", int(time.Now().Unix()), labelKeys, labelValues, "request last update for this IP")
			}

		}

	}

}

func printLog(msg ...interface{}) {
	if debugFlag {
		fmt.Println(msg...)
	}
}

func systemAlive(listenAddress, metricsPath string) {

	// systemd alive check
	var metricsURL string
	if !strings.HasPrefix(listenAddress, ":") {
		// User has provided address + port
		metricsURL = "http://" + listenAddress + metricsPath
	} else {
		// User has provided :port only - we need to check ourselves on 127.0.0.1
		metricsURL = "http://127.0.0.1" + listenAddress + metricsPath
	}

	// Call the metrics URL...
	res, err := http.Get(metricsURL)
	if err == nil {
		// ... and notify systemd that everything was ok
		daemon.SdNotify(false, daemon.SdNotifyWatchdog)
	} else {
		// ... do nothing if it was not ok, but log. Systemd will restart soon.
		log.Println("ERR: liveness check failed: ", err)
	}

	// Read all away to free memory
	_, _ = io.ReadAll(res.Body)
	defer res.Body.Close()

}

func setPrometheusMetric(key string, value int, labels []string, labelValues []string, help string) {

	// Check if metric is already registered, if not, register it
	_, ok := exposed[key]
	if !ok {
		exposed[key] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "dd_" + key,
			Help: help,
		},
			labels,
		)

		prometheus.MustRegister(exposed[key])
	}

	// Now set the value
	exposed[key].WithLabelValues(labelValues...).Set(float64(value))

	// Update lastUpdate so we immediately see when no updates happen anymore
	lastUpdate.WithLabelValues("global").Set(float64(time.Now().Unix()))

}
