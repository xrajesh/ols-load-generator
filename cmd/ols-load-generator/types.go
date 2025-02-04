package main

import (
	"time"
)

// Type to store the test config.
type TestConfig struct {
	Host       string        `json:"host"`
	Duration   time.Duration `json:"duration"`
	Workers    int           `json:"workers"`
	AuthToken  string        `json:"-"`
	Uuid       string        `json:"uuid"`
	ESHost     string        `json:"eshost"`
	ESIndex    string        `json:"esindex"`
	MetricStep int           `json:"metricstep"`
	Profiles   []string      `json:"profiles"`
	QueryOnly  bool          `json:"queryonly"`
}
