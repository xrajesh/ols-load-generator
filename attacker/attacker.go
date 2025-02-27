package attacker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloud-bulldozer/go-commons/indexers"
	"github.com/quay/zlog"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const requestTimeout = 120

// generateVegetaRequests generates requests which can be fed as input to vegeta for HTTP benchmarking.
// It return a consolidated targets list which has all the requests fed all at once to vegeta.
func generateVegetaRequests(requestDicts []map[string]interface{}) []vegeta.Target {
	// Convert requestDicts to a slice of Vegeta requests
	var targets []vegeta.Target
	for _, reqDict := range requestDicts {
		var req_body []byte
		var req_headers http.Header
		// Prepare request body
		if reqDict["body"] != nil && reqDict["method"] != http.MethodGet {
			req_body, _ = reqDict["body"].([]byte)
		}
		// Set the request headers
		if headers, ok := reqDict["header"]; ok {
			req_headers = headers.(http.Header)
		} else {
			req_headers = http.Header{
				"Authorization": []string{"Bearer " + os.Getenv("AUTH_TOKEN")},
				"Content-Type":  []string{"application/json"},
				"accept":        []string{"application/json"},
			}
		}
		// Vegeta Target
		target := vegeta.Target{
			Method: reqDict["method"].(string),
			URL:    reqDict["url"].(string),
			Header: req_headers,
			Body:   req_body,
		}
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		panic("Something is wrong with requests. Requests list cannot be empty")
	}
	return targets
}

// indexVegetaResults to process vegeta output and index the results to elastic search.
// It returns an error if any during the execution.
func indexVegetaResults(ctx context.Context, metrics vegeta.Metrics, testName string, attackMap map[string]string, startTime, endTime time.Time) error {
	var indexer *indexers.Indexer
	var indexerConfig indexers.IndexerConfig

	if attackMap["ESHost"] != "" && attackMap["ESIndex"] != "" {
		zlog.Info(ctx).Msg("Creating opensearch indexer")
		indexerConfig = indexers.IndexerConfig{
			Type:               indexers.OpenSearchIndexer,
			Servers:            []string{attackMap["ESHost"]},
			Index:              attackMap["ESIndex"],
			InsecureSkipVerify: true,
		}
		zlog.Info(ctx).Str("es-index", indexerConfig.Index).Msg("Indexing documents")
	} else {
		zlog.Info(ctx).Msg("Creating local indexer")
		indexerConfig = indexers.IndexerConfig{
			Type:             indexers.LocalIndexer,
			MetricsDirectory: "collected-metrics" + "-" + attackMap["Uuid"],
		}
		zlog.Info(ctx).Str("metrics-directory", indexerConfig.MetricsDirectory).Msg("Indexing documents")
	}
	indexer, err := indexers.NewIndexer(indexerConfig)
	if err != nil {
		return fmt.Errorf("Failure while creating an indexer: %w", err)
	}
	workers, _ := strconv.Atoi(attackMap["Workers"])
	queryOnly, _ := strconv.ParseBool(attackMap["QueryOnly"])
	hostname, _ := os.Hostname()
	resp, err := (*indexer).Index([]interface{}{Document{
		Workload:       "ols-load-generator",
		Endpoint:       attackMap["Host"],
		RequestTimeout: requestTimeout,
		MetricName:     testName,
		Hostname:       hostname,
		Duration:       attackMap["Duration"],
		Workers:        workers,
		AttackTime:     metrics.Duration,
		WaitTime:       metrics.Wait,
		Throughput:     metrics.Throughput,
		StatusCodes:    metrics.StatusCodes,
		Requests:       metrics.Requests,
		P99Latency:     metrics.Latencies.P99,
		P95Latency:     metrics.Latencies.P95,
		MaxLatency:     metrics.Latencies.Max,
		MinLatency:     metrics.Latencies.Min,
		ReqLatency:     metrics.Latencies.Mean,
		Timestamp:      startTime.UTC().Format("2006-01-02T15:04:05.999999999Z"),
		EndTimestamp:   endTime.UTC().Format("2006-01-02T15:04:05.999999999Z"),
		ElapsedTime:    endTime.UTC().Sub(startTime.UTC()).Round(time.Second).Seconds(),
		BytesIn:        metrics.BytesIn.Mean,
		BytesOut:       metrics.BytesOut.Mean,
		Uuid:           attackMap["Uuid"],
		QueryOnly:      queryOnly,
	}}, indexers.IndexingOpts{
		MetricName: testName,
	})
	if err != nil {
		return err
	}
	zlog.Info(ctx).Msg(resp)
	return nil
}

// RunVegeta runs vegeta, records their results and indexes to elastic search if provided with connection details.
// It returns an error if any during the execution.
func RunVegeta(ctx context.Context, requestDicts []map[string]interface{}, testName string, attackMap map[string]string) error {
	startTime := time.Now()
	requests := generateVegetaRequests(requestDicts)
	workers, _ := strconv.Atoi(attackMap["Workers"])
	duration, err := time.ParseDuration(attackMap["Duration"])
	if err != nil {
		return fmt.Errorf("Failed to parse duration: %v", err)
	}
	rate := vegeta.Rate{Freq: 0, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(requests...)
	// Custom HTTP client with InsecureSkipVerify set to true
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	attacker := vegeta.NewAttacker(vegeta.Client(client),
		vegeta.Workers(uint64(workers)),
		vegeta.MaxWorkers(uint64(workers)),
		vegeta.Connections(workers),
		vegeta.MaxConnections(workers),
		vegeta.KeepAlive(true),
		vegeta.Timeout(requestTimeout*time.Second))

	// Initiate vegeta attack and stop immediately after completion
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Vegeta Attack") {
		metrics.Add(res)
	}

	metrics.Close()

	// Generate Vegeta text report
	report := vegeta.NewTextReporter(&metrics)
	err = report.Report(os.Stdout)
	if err != nil {
		return fmt.Errorf("vegeta report command failure: %w", err)
	}
	zlog.Info(ctx).Msg("Vegeta attack completed successfully")
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	zlog.Info(ctx).Stringer("duration", elapsedTime).Msg(fmt.Sprintf("Total time taken for %s", testName))

	err = indexVegetaResults(ctx, metrics, testName, attackMap, startTime, endTime)
	if err != nil {
		return fmt.Errorf("Failed to index vegeta results: %w", err)
	}

	return nil
}
