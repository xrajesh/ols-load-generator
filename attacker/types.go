package attacker

import (
	"time"
)

// Type used to index results to elastic search.
type Document struct {
	Workload       string         `json:"workload"`
	Endpoint       string         `json:"endpoint"`
	RequestTimeout int            `json:"requestTimeout"`
	MetricName     string         `json:"metricName"`
	Hostname       string         `json:"hostname"`
	Duration       string         `json:"duration"`
	Workers        int            `json:"workers"`
	AttackTime     time.Duration  `json:"attackTime"`
	WaitTime       time.Duration  `json:"waitTime"`
	Throughput     float64        `json:"throughput"`
	StatusCodes    map[string]int `json:"statusCodes"`
	Requests       uint64         `json:"requests"`
	P99Latency     time.Duration  `json:"p99Latency"`
	P95Latency     time.Duration  `json:"p95Latency"`
	MaxLatency     time.Duration  `json:"maxLatency"`
	MinLatency     time.Duration  `json:"minLatency"`
	ReqLatency     time.Duration  `json:"reqLatency"`
	Timestamp      string         `json:"timestamp"`
	EndTimestamp   string         `json:"endTimestamp"`
	ElapsedTime    float64        `json:"elapsedTime"`
	BytesIn        float64        `json:"bytesIn"`
	BytesOut       float64        `json:"bytesOut"`
	Uuid           string         `json:"uuid"`
	QueryOnly      bool           `json:"queryonly"`
}
