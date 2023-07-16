package main

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/spacelift-io/prometheus-exporter/client"
	"github.com/spacelift-io/prometheus-exporter/client/session"
	"github.com/spacelift-io/prometheus-exporter/logging"
)

type spaceliftCollector struct {
	ctx                                    context.Context
	logger                                 *zap.SugaredLogger
	client                                 client.Client
	scrapeTimeout                          time.Duration
	publicRunsPending                      *prometheus.Desc
	publicWorkersBusy                      *prometheus.Desc
	publicParallelism                      *prometheus.Desc
	workerPoolRunsPending                  *prometheus.Desc
	workerPoolWorkersBusy                  *prometheus.Desc
	workerPoolWorkers                      *prometheus.Desc
	workerPoolWorkersDrained               *prometheus.Desc
	currentBillingPeriodStart              *prometheus.Desc
	currentBillingPeriodEnd                *prometheus.Desc
	currentBillingPeriodUsedPrivateSeconds *prometheus.Desc
	currentBillingPeriodUsedPublicSeconds  *prometheus.Desc
	currentBillingPeriodUsedSeats          *prometheus.Desc
	StackStateDiscarded                    *prometheus.Desc
	StackStateFailed                       *prometheus.Desc
	StackStateFinished                     *prometheus.Desc
	StackStateNone                         *prometheus.Desc
	totalStacksCount                       *prometheus.Desc
	scrapeDuration                         *prometheus.Desc
	buildInfo                              *prometheus.Desc
}

func newSpaceliftCollector(ctx context.Context, httpClient *http.Client, session session.Session, scrapeTimeout time.Duration) (prometheus.Collector, error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("could not read build info")
	}

	return &spaceliftCollector{
		ctx:           ctx,
		logger:        logging.FromContext(ctx).Sugar(),
		client:        client.New(httpClient, session),
		scrapeTimeout: scrapeTimeout,
		StackStateDiscarded: prometheus.NewDesc(
			"Stack_State_DISCARDED",
			"Stack state Discarded",
			[]string{"stack_name"},
			nil),
		StackStateFailed: prometheus.NewDesc(
			"Stack_State_FAILED",
			"Stack state FAILED",
			[]string{"stack_name"},
			nil),
		StackStateFinished: prometheus.NewDesc(
			"Stack_State_FINISHED",
			"Stack state FINISHED",
			[]string{"stack_name"},
			nil),
		StackStateNone: prometheus.NewDesc(
			"Stack_State_NONE",
			"Stack state NONE",
			[]string{"stack_name"},
			nil),
		totalStacksCount: prometheus.NewDesc(
			"total_stacks_count",
			"Total Stacks Count",
			nil,
			nil),
		publicRunsPending: prometheus.NewDesc(
			"spacelift_public_worker_pool_runs_pending",
			"The number of runs in your account currently queued and waiting for a public worker",
			nil,
			nil),
		publicWorkersBusy: prometheus.NewDesc(
			"spacelift_public_worker_pool_workers_busy",
			"The number of currently busy workers in the public worker pool for this account",
			nil,
			nil),
		publicParallelism: prometheus.NewDesc(
			"spacelift_public_worker_pool_parallelism",
			"The maximum number of simultaneously executing runs on the public worker pool for this account",
			nil,
			nil),
		workerPoolRunsPending: prometheus.NewDesc(
			"spacelift_worker_pool_runs_pending",
			"The number of runs currently queued and waiting for a worker from a particular pool",
			[]string{"worker_pool_id", "worker_pool_name"},
			nil),
		workerPoolWorkersBusy: prometheus.NewDesc(
			"spacelift_worker_pool_workers_busy",
			"The number of currently busy workers in a worker pool",
			[]string{"worker_pool_id", "worker_pool_name"},
			nil),
		workerPoolWorkers: prometheus.NewDesc(
			"spacelift_worker_pool_workers",
			"The number of workers in a worker pool",
			[]string{"worker_pool_id", "worker_pool_name"},
			nil),
		workerPoolWorkersDrained: prometheus.NewDesc(
			"spacelift_worker_pool_workers_drained",
			"The number of workers in a worker pool that have been drained",
			[]string{"worker_pool_id", "worker_pool_name"},
			nil),
		currentBillingPeriodStart: prometheus.NewDesc(
			"spacelift_current_billing_period_start_timestamp_seconds",
			"The timestamp of the start of the current billing period",
			nil,
			nil),
		currentBillingPeriodEnd: prometheus.NewDesc(
			"spacelift_current_billing_period_end_timestamp_seconds",
			"The timestamp of the end of the current billing period",
			nil,
			nil),
		currentBillingPeriodUsedPrivateSeconds: prometheus.NewDesc(
			"spacelift_current_billing_period_used_private_seconds",
			"The amount of private worker usage in the current billing period",
			nil,
			nil),
		currentBillingPeriodUsedPublicSeconds: prometheus.NewDesc(
			"spacelift_current_billing_period_used_public_seconds",
			"The amount of public worker usage in the current billing period",
			nil,
			nil),
		currentBillingPeriodUsedSeats: prometheus.NewDesc(
			"spacelift_current_billing_period_used_seats",
			"The number of seats used in the current billing period",
			nil,
			nil),
		scrapeDuration: prometheus.NewDesc(
			"spacelift_scrape_duration_seconds",
			"The duration in seconds of the request to the Spacelift API for metrics",
			nil,
			nil),
		buildInfo: prometheus.NewDesc(
			"spacelift_build_info",
			"Contains build information about the exporter",
			nil,
			prometheus.Labels{"version": version, "commit": commit, "goversion": buildInfo.GoVersion}),
	}, nil
}

func (c *spaceliftCollector) Describe(descriptorChannel chan<- *prometheus.Desc) {
	descriptorChannel <- c.publicRunsPending
	descriptorChannel <- c.publicWorkersBusy
	descriptorChannel <- c.publicParallelism
	descriptorChannel <- c.workerPoolRunsPending
	descriptorChannel <- c.workerPoolWorkersBusy
	descriptorChannel <- c.workerPoolWorkersDrained
	descriptorChannel <- c.currentBillingPeriodStart
	descriptorChannel <- c.currentBillingPeriodEnd
	descriptorChannel <- c.currentBillingPeriodUsedPrivateSeconds
	descriptorChannel <- c.currentBillingPeriodUsedPublicSeconds
	descriptorChannel <- c.currentBillingPeriodUsedSeats
	descriptorChannel <- c.StackStateDiscarded
	descriptorChannel <- c.StackStateFailed
	descriptorChannel <- c.StackStateFinished
	descriptorChannel <- c.StackStateNone
	descriptorChannel <- c.totalStacksCount
	descriptorChannel <- c.buildInfo
}

type metricsQuery struct {
	PublicWorkerPool struct {
		Parallelism int `graphql:"parallelism"`
		BusyWorkers int `graphql:"busyWorkers"`
		PendingRuns int `graphql:"pendingRuns"`
	} `graphql:"publicWorkerPool"`
	WorkerPools []struct {
		ID          string `graphql:"id"`
		Name        string `graphql:"name"`
		PendingRuns int    `graphql:"pendingRuns"`
		BusyWorkers int    `graphql:"busyWorkers"`
		Workers     []struct {
			ID      string `graphql:"id"`
			Drained bool   `graphql:"drained"`
		} `graphql:"workers"`
	} `graphql:"workerPools"`
	Usage struct {
		BillingPeriodStart int `graphql:"billingPeriodStart"`
		BillingPeriodEnd   int `graphql:"billingPeriodEnd"`
		UsedPrivateMinutes int `graphql:"usedPrivateMinutes"`
		UsedPublicMinutes  int `graphql:"usedPublicMinutes"`
		UsedSeats          int `graphql:"usedSeats"`
	} `graphql:"usage"`
	Stacks []struct {
		ID             string `graphql:"id"`
		Name           string `graphql:"name"`
		Administrative bool   `graphql:"administrative"`
		CreatedAt      int    `graphql:"createdAt"`
		Description    string `graphql:"description"`
		State          string `graphql:"state"`
		Runs           []struct {
			IsMostRecent bool `graphql:"isMostRecent"`
			Finished     bool `graphql:"finished"`
		} `graphql:"runs"`
	} `graphql:"stacks"`
}

func (c *spaceliftCollector) Collect(metricChannel chan<- prometheus.Metric) {
	var query metricsQuery

	start := time.Now()
	err := func() error {
		ctx, cancel := context.WithTimeout(c.ctx, c.scrapeTimeout)
		defer cancel()
		return c.client.Query(ctx, &query, nil)
	}()

	scrapeDuration := time.Since(start)
	metricChannel <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, scrapeDuration.Seconds())

	if err != nil {
		msg := "Failed to request metrics from the Spacelift API"
		if err == context.DeadlineExceeded {
			msg = "The request to the Spacelift API for metric data timed out"
		}

		c.logger.Errorw(msg, zap.Error(err))
		metricChannel <- prometheus.NewInvalidMetric(
			prometheus.NewDesc(
				"spacelift_error",
				msg,
				nil,
				nil),
			err)

		return
	}

	metricChannel <- prometheus.MustNewConstMetric(c.buildInfo, prometheus.GaugeValue, 1)
	metricChannel <- prometheus.MustNewConstMetric(c.publicRunsPending, prometheus.GaugeValue, float64(query.PublicWorkerPool.PendingRuns))
	metricChannel <- prometheus.MustNewConstMetric(c.publicWorkersBusy, prometheus.GaugeValue, float64(query.PublicWorkerPool.BusyWorkers))
	metricChannel <- prometheus.MustNewConstMetric(c.publicParallelism, prometheus.GaugeValue, float64(query.PublicWorkerPool.Parallelism))
	metricChannel <- prometheus.MustNewConstMetric(c.currentBillingPeriodStart, prometheus.GaugeValue, float64(query.Usage.BillingPeriodStart))
	metricChannel <- prometheus.MustNewConstMetric(c.currentBillingPeriodEnd, prometheus.GaugeValue, float64(query.Usage.BillingPeriodEnd))
	metricChannel <- prometheus.MustNewConstMetric(c.currentBillingPeriodUsedPrivateSeconds, prometheus.GaugeValue, float64(query.Usage.UsedPrivateMinutes*60))
	metricChannel <- prometheus.MustNewConstMetric(c.currentBillingPeriodUsedPublicSeconds, prometheus.GaugeValue, float64(query.Usage.UsedPublicMinutes*60))
	metricChannel <- prometheus.MustNewConstMetric(c.currentBillingPeriodUsedSeats, prometheus.GaugeValue, float64(query.Usage.UsedSeats))

	for _, stack := range query.Stacks {
		isDiscarded := stack.State == "DISCARDED"
		isFailed := stack.State == "FAILED"
		isFinished := stack.State == "FINISHED"
		isNone := stack.State == "NONE"
		if isDiscarded {
			discarded := 1
			metricChannel <- prometheus.MustNewConstMetric(c.StackStateDiscarded, prometheus.GaugeValue, float64(discarded), stack.Name)
		}
		if isFailed {
			discarded := 1
			metricChannel <- prometheus.MustNewConstMetric(c.StackStateFailed, prometheus.GaugeValue, float64(discarded), stack.Name)
		}
		if isFinished {
			discarded := 1
			metricChannel <- prometheus.MustNewConstMetric(c.StackStateFinished, prometheus.GaugeValue, float64(discarded), stack.Name)
		}
		if isNone {
			finished := 1
			metricChannel <- prometheus.MustNewConstMetric(c.StackStateNone, prometheus.GaugeValue, float64(finished), stack.Name)
		}
	}
	metricChannel <- prometheus.MustNewConstMetric(c.totalStacksCount, prometheus.GaugeValue, float64(len(query.Stacks)))

	for _, workerPool := range query.WorkerPools {
		metricChannel <- prometheus.MustNewConstMetric(c.workerPoolRunsPending, prometheus.GaugeValue, float64(workerPool.PendingRuns), workerPool.ID, workerPool.Name)
		metricChannel <- prometheus.MustNewConstMetric(c.workerPoolWorkersBusy, prometheus.GaugeValue, float64(workerPool.BusyWorkers), workerPool.ID, workerPool.Name)
		metricChannel <- prometheus.MustNewConstMetric(c.workerPoolWorkers, prometheus.GaugeValue, float64(len(workerPool.Workers)), workerPool.ID, workerPool.Name)

		drained := 0
		for _, worker := range workerPool.Workers {
			if worker.Drained {
				drained++
			}
		}
		metricChannel <- prometheus.MustNewConstMetric(c.workerPoolWorkersDrained, prometheus.GaugeValue, float64(drained), workerPool.ID, workerPool.Name)
	}
}
