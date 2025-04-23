package queue

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleMetrics = `
# HELP test_metric This is a test metric
# TYPE test_metric gauge
test_metric 123.45
`

func TestFetchMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleMetrics)
	}))
	defer server.Close()

	metrics, err := fetchMetrics(server.URL)
	assert.NoError(t, err)
	assert.Contains(t, metrics, "test_metric")
}

func TestExtractMetricValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleMetrics)
	}))
	defer server.Close()

	metrics, _ := fetchMetrics(server.URL)
	value := extractMetricValue(metrics, "test_metric")
	assert.Equal(t, 123.45, value)
}

func TestExtractMetricValue_MissingMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sampleMetrics)
	}))
	defer server.Close()

	metrics, _ := fetchMetrics(server.URL)
	value := extractMetricValue(metrics, "non_existent_metric")
	assert.Equal(t, 0.0, value)
}
