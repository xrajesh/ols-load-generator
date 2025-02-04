package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openshift/ols-load-generator/attacker"
	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
)

// Command line to handle attack functionality.
var AttackCmd = &cli.Command{
	Name:        "attack",
	Description: "perform attack on ols endpoints",
	Usage:       "ols-load-generator attack",
	Action:      attackAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Usage:   "--host localhost:6060",
			Value:   "http://localhost:6060",
			EnvVars: []string{"OLS_TEST_HOST"},
		},
		&cli.StringFlag{
			Name:    "authtoken",
			Usage:   "--authtoken authtoken",
			Value:   "",
			EnvVars: []string{"OLS_TEST_AUTH_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "uuid",
			Usage:   "--uuid f519d9b2-aa62-44ab-9ce8-4156b712f6d2",
			Value:   uuid.New().String(),
			EnvVars: []string{"OLS_TEST_UUID"},
		},
		&cli.DurationFlag{
			Name:    "duration",
			Usage:   "--duration 1m",
			Value:   1 * time.Minute,
			EnvVars: []string{"OLS_TEST_DURATION"},
		},
		&cli.IntFlag{
			Name:    "workers",
			Usage:   "--workers 10",
			Value:   10,
			EnvVars: []string{"OLS_TEST_WORKERS"},
		},
		&cli.StringFlag{
			Name:    "eshost",
			Usage:   "--eshost eshosturl",
			Value:   "",
			EnvVars: []string{"OLS_TEST_ES_HOST"},
		},
		&cli.StringFlag{
			Name:    "esindex",
			Usage:   "--esindex esindex",
			Value:   "",
			EnvVars: []string{"OLS_TEST_ES_INDEX"},
		},
		&cli.IntFlag{
			Name:    "metricstep",
			Usage:   "--metricstep 30",
			Value:   30,
			EnvVars: []string{"OLS_TEST_METRIC_STEP"},
		},
		&cli.StringFlag{
			Name:    "profiles",
			Usage:   "--profiles metrics.yaml,metrics-report.yaml",
			Value:   "attacker/assets/profiles/metrics-report.yaml,attacker/assets/profiles/metrics-timeseries.yaml",
			EnvVars: []string{"OLS_TEST_PROFILES"},
		},
		&cli.BoolFlag{
			Name:    "queryonly",
			Usage:   "--query",
			Value:   false,
			EnvVars: []string{"OLS_TEST_QUERY_ONLY"},
		},
	},
}

// attackConfig creates and returns a test configuration from CLI options.
func attackConfig(c *cli.Context) *TestConfig {
	profilesArg := c.String("profiles")
	return &TestConfig{
		AuthToken:  c.String("authtoken"),
		Uuid:       c.String("uuid"),
		Host:       c.String("host"),
		Duration:   c.Duration("duration"),
		Workers:    c.Int("workers"),
		ESHost:     c.String("eshost"),
		ESIndex:    c.String("esindex"),
		MetricStep: c.Int("metricstep"),
		QueryOnly:  c.Bool("queryonly"),
		Profiles:   strings.Split(strings.TrimSpace(profilesArg), ","),
	}
}

// attackAction drives the attacking logic.
// It returns an error if any during the execution.
func attackAction(c *cli.Context) error {
	startTime := time.Now()
	ctx := c.Context
	conf := attackConfig(c)

	zlog.Info(ctx).Msg("🔥 Orchestrating the workload")
	err := orchestrateWorkload(ctx, conf)
	if err != nil {
		return err
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	err = IndexMetrics(ctx, conf, startTime.Add(-10*time.Minute), endTime.Add(10*time.Minute))
	if err != nil {
		return err
	}
	zlog.Info(ctx).Stringer("duration", elapsedTime).Msg("Total time taken for completion")
	return nil
}

// orchestrateWorkload triggers the api endpoint hits and writes results to the desired location.
// It returns an error if any during the execution.
func orchestrateWorkload(ctx context.Context, conf *TestConfig) error {
	zlog.Info(ctx).Str("Uuid", conf.Uuid).Msg("Run details")
	var requests []map[string]interface{}
	var err error
	attackMap := map[string]string{
		"Uuid":      conf.Uuid,
		"Workers":   strconv.Itoa(conf.Workers),
		"Duration":  conf.Duration.String(),
		"ESHost":    conf.ESHost,
		"ESIndex":   conf.ESIndex,
		"Host":      conf.Host,
		"QueryOnly": strconv.FormatBool(conf.QueryOnly),
	}

	if !conf.QueryOnly {
		requests = attacker.CreateReadinessRequests(ctx, conf.Duration, conf.Workers, conf.Host)
		err = attacker.RunVegeta(ctx, requests, "get_readiness", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running GET operation on /readiness: %w", err)
		}

		requests = attacker.CreateLivenessRequests(ctx, conf.Duration, conf.Workers, conf.Host)
		err = attacker.RunVegeta(ctx, requests, "get_liveness", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running GET operation on /liveness: %w", err)
		}

		requests = attacker.CreateAuthorizedRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken)
		err = attacker.RunVegeta(ctx, requests, "post_authorized", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running POST operation on /authorized: %w", err)
		}

		requests = attacker.CreateMetricsRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken)
		err = attacker.RunVegeta(ctx, requests, "get_metrics", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running GET operation on /metrics: %w", err)
		}

		requests = attacker.CreateGetFeedbackStatusRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken)
		err = attacker.RunVegeta(ctx, requests, "get_feedback_status", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running GET operation on /v1/feedback/status: %w", err)
		}

		requests = attacker.CreateFeedbackRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken)
		err = attacker.RunVegeta(ctx, requests, "post_feedback", attackMap)
		if err != nil {
			return fmt.Errorf("Error while running POST operation on /v1/feedback: %w", err)
		}
	}

	requests = attacker.CreateQueryRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken, false, false)
	err = attacker.RunVegeta(ctx, requests, "post_query", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running POST operation on /v1/query: %w", err)
	}

	requests = attacker.CreateQueryRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken, false, true)
	err = attacker.RunVegeta(ctx, requests, "post_streaming_query", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running POST operation on /v1/streaming_query: %w", err)
	}

	requests = attacker.CreateQueryRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken, true, false)
	err = attacker.RunVegeta(ctx, requests, "post_query_with_cache", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running POST operation on /v1/query with cache: %w", err)
	}

	requests = attacker.CreateQueryRequests(ctx, conf.Duration, conf.Workers, conf.Host, conf.AuthToken, true, true)
	err = attacker.RunVegeta(ctx, requests, "post_streaming_query_with_cache", attackMap)
	if err != nil {
		return fmt.Errorf("Error while running POST operation on /v1/streaming_query with cache: %w", err)
	}

	zlog.Info(ctx).Str("Uuid", conf.Uuid).Msg("👋 Exiting ols-load-generator")
	return nil
}
