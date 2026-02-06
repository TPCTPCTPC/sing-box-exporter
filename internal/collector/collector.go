package collector

import (
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/TPCTPCTPC/sing-box-exporter/internal/v2ray"
)

const (
	namespace = "singbox"
)

type SingBoxCollector struct {
	client      *v2ray.Client
	users       []string
	inbounds    []string
	
	// Metrics Descriptors
	upBytes     *prometheus.Desc
	downBytes   *prometheus.Desc
	userUp      *prometheus.Desc
	userDown    *prometheus.Desc
	scrapeDur   *prometheus.Desc
	scrapeOk    *prometheus.Desc
}

func NewSingBoxCollector(client *v2ray.Client, users []string, inbounds []string) *SingBoxCollector {
	return &SingBoxCollector{
		client:   client,
		users:    users,
		inbounds: inbounds,
		
		upBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "inbound", "traffic_uplink_bytes"),
			"Uplink traffic for an inbound",
			[]string{"inbound"}, nil,
		),
		downBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "inbound", "traffic_downlink_bytes"),
			"Downlink traffic for an inbound",
			[]string{"inbound"}, nil,
		),
		userUp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "user", "traffic_uplink_bytes"),
			"Uplink traffic for a user",
			[]string{"user"}, nil,
		),
		userDown: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "user", "traffic_downlink_bytes"),
			"Downlink traffic for a user",
			[]string{"user"}, nil,
		),
		scrapeDur: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "exporter", "scrape_duration_seconds"),
			"Duration of the scrape",
			nil, nil,
		),
		scrapeOk: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "exporter", "scrape_success"),
			"Whether the scrape was successful",
			nil, nil,
		),
	}
}

func (c *SingBoxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upBytes
	ch <- c.downBytes
	ch <- c.userUp
	ch <- c.userDown
	ch <- c.scrapeDur
	ch <- c.scrapeOk
}

func (c *SingBoxCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	success := 1.0

	stats, err := c.client.QueryAllStats()
	if err != nil {
		log.Printf("[Error] Failed to scrape V2Ray stats: %v", err)
		success = 0.0
	} else {
		log.Printf("[Debug] Fetched %d stats from sing-box", len(stats))
		// Process all stats
		for _, s := range stats {
			// Name format: type>>>id>>>...
			parts := strings.Split(s.Name, ">>>")
			if len(parts) < 4 {
				continue
			}
			
			// parts[0] = "user" or "inbound"
			// parts[1] = identifier (email or tag)
			// parts[2] = "traffic"
			// parts[3] = "uplink" or "downlink"
			
			metricType := parts[0]
			identifier := parts[1]
			direction := parts[3]
			
			// Filter based on initial config if provided, otherwise allow all
			// (If c.users is empty, we allow all users. Same for inbounds)
			if metricType == "user" {
				if len(c.users) > 0 && !contains(c.users, identifier) { continue }
				
				if direction == "uplink" {
					ch <- prometheus.MustNewConstMetric(c.userUp, prometheus.CounterValue, float64(s.Value), identifier)
				} else if direction == "downlink" {
					ch <- prometheus.MustNewConstMetric(c.userDown, prometheus.CounterValue, float64(s.Value), identifier)
				}
			} else if metricType == "inbound" {
				if len(c.inbounds) > 0 && !contains(c.inbounds, identifier) { continue }

				if direction == "uplink" {
					ch <- prometheus.MustNewConstMetric(c.upBytes, prometheus.CounterValue, float64(s.Value), identifier)
				} else if direction == "downlink" {
					ch <- prometheus.MustNewConstMetric(c.downBytes, prometheus.CounterValue, float64(s.Value), identifier)
				}
			}
		}
	}
	
	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDur, prometheus.GaugeValue, duration)
	ch <- prometheus.MustNewConstMetric(c.scrapeOk, prometheus.GaugeValue, success)
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item { return true }
    }
    return false
}
