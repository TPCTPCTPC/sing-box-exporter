package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/sing-box-exporter/internal/collector"
	"github.com/yourusername/sing-box-exporter/internal/v2ray"
)

var (
	listenAddr  = flag.String("listen", ":9091", "Address to listen on for web interface and telemetry.")
	sbAddr      = flag.String("singbox", "127.0.0.1:19998", "Address of the sing-box V2Ray stats API.")
	users       = flag.String("users", "", "Comma-separated list of users to monitor (e.g. 'user1,user2').")
	inbounds    = flag.String("inbounds", "main,relay,route", "Comma-separated list of inbounds to monitor.")
	metricsPath = flag.String("telemetry-path", "/metrics", "Path under which to expose metrics.")
)

func main() {
	flag.Parse()

	if *users == "" {
		log.Println("Warning: No users specified to monitor. Use -users flag.")
	}

	userList := strings.Split(*users, ",")
	inboundList := strings.Split(*inbounds, ",")
	
	// Clean up lists
	for i := range userList { userList[i] = strings.TrimSpace(userList[i]) }
	for i := range inboundList { inboundList[i] = strings.TrimSpace(inboundList[i]) }

	// Initialize V2Ray Client
	vClient, err := v2ray.NewClient(*sbAddr, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to create V2Ray client: %v", err)
	}
	// Note: We don't defer vClient.Close() here effectively because we run forever,
	// but strictly speaking main termination closes it.

	// Register Collector
	exporter := collector.NewSingBoxCollector(vClient, userList, inboundList)
	prometheus.MustRegister(exporter)

	// Start HTTP Server
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Sing-box Exporter</title></head>
             <body>
             <h1>Sing-box Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Printf("Starting sing-box exporter on %s", *listenAddr)
	log.Printf("Monitoring sing-box at %s", *sbAddr)
	log.Printf("Monitoring users: %v", userList)
	
	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("Error starting web server: %v", err)
	}
}
