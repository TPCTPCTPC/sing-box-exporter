package collector

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yourusername/sing-box-exporter/internal/v2ray"
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
	
	var wg sync.WaitGroup
	
	// Collect Inbound Stats
	for _, inbound := range c.inbounds {
		if inbound == "" { continue }
		wg.Add(1)
		go func(tag string) {
			defer wg.Done()
			up, down, err := c.client.GetInboundStats(tag)
			if err != nil {
				// We don't log every miss as it spams, but maybe we should log fatal errors?
				// For now, silent fail on specific metrics is better than breaking the whole scrape
			}
			ch <- prometheus.MustNewConstMetric(c.upBytes, prometheus.CounterValue, float64(up), tag)
			ch <- prometheus.MustNewConstMetric(c.downBytes, prometheus.CounterValue, float64(down), tag)
		}(inbound)
	}

	// Collect User Stats
	for _, user := range c.users {
		if user == "" { continue }
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			up, down, err := c.client.GetUserStats(email)
			if err != nil {
				// Again, silent or debug log
			}
			ch <- prometheus.MustNewConstMetric(c.userUp, prometheus.CounterValue, float64(up), email)
			ch <- prometheus.MustNewConstMetric(c.userDown, prometheus.CounterValue, float64(down), email)
		}(user)
	}
	
	wg.Wait()
	
	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDur, prometheus.GaugeValue, duration)
	ch <- prometheus.MustNewConstMetric(c.scrapeOk, prometheus.GaugeValue, success)
}
