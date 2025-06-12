package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"
)

const (
	MetricKeyConnections        = "connections"
	MetricKeySubscriptions      = "subscriptions"
	MetricKeyMessagesSent       = "messages_sent"
	MetricKeyBufferOverflow     = "buffer_overflow"
	MetricKeyConnectionDuration = "connection_duration"
	MetricKeyConnectionRejected = "connection_rejected"
	MetricKeyCircuitBreakerOpen = "circuit_breaker_open"
	MetricsModuleName           = "stream"
)

// UpdateConnectionMetrics updates connection-related metrics
func UpdateConnectionMetrics(totalConnections int32) {
	telemetry.SetGaugeWithLabels(
		[]string{MetricsModuleName, MetricKeyConnections},
		float32(totalConnections),
		[]metrics.Label{},
	)
}

// UpdateSubscriptionMetrics updates subscription-related metrics
func UpdateSubscriptionMetrics(subscriptionType string, count int) {
	telemetry.SetGaugeWithLabels(
		[]string{MetricsModuleName, MetricKeySubscriptions},
		float32(count),
		[]metrics.Label{
			{Name: "type", Value: subscriptionType},
		},
	)
}

// IncrementMessagesSent increments the messages sent counter
func IncrementMessagesSent() {
	telemetry.IncrCounter(
		1,
		MetricsModuleName+"_"+MetricKeyMessagesSent,
	)
}

// IncrementBufferOverflow increments the buffer overflow counter
func IncrementBufferOverflow(subscriptionType string) {
	telemetry.IncrCounterWithLabels(
		[]string{MetricsModuleName, MetricKeyBufferOverflow},
		1,
		[]metrics.Label{
			{Name: "subscription_type", Value: subscriptionType},
		},
	)
}

// RecordConnectionDuration records the duration of a connection
func RecordConnectionDuration(startTime time.Time) {
	telemetry.ModuleMeasureSince(MetricsModuleName, startTime, MetricKeyConnectionDuration)
}

// IncrementConnectionRejected increments the connection rejected counter
func IncrementConnectionRejected(reason string) {
	telemetry.IncrCounterWithLabels(
		[]string{MetricsModuleName, MetricKeyConnectionRejected},
		1,
		[]metrics.Label{
			{Name: "reason", Value: reason},
		},
	)
}

// UpdateCircuitBreakerMetrics updates circuit breaker metrics
func UpdateCircuitBreakerMetrics(openCircuits, halfOpenCircuits, closedCircuits int) {
	telemetry.SetGaugeWithLabels(
		[]string{MetricsModuleName, MetricKeyCircuitBreakerOpen},
		float32(openCircuits),
		[]metrics.Label{
			{Name: "state", Value: "open"},
		},
	)
	telemetry.SetGaugeWithLabels(
		[]string{MetricsModuleName, MetricKeyCircuitBreakerOpen},
		float32(halfOpenCircuits),
		[]metrics.Label{
			{Name: "state", Value: "half_open"},
		},
	)
	telemetry.SetGaugeWithLabels(
		[]string{MetricsModuleName, MetricKeyCircuitBreakerOpen},
		float32(closedCircuits),
		[]metrics.Label{
			{Name: "state", Value: "closed"},
		},
	)
}
