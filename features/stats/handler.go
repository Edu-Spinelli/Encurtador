package stats

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/spinelli/encurtador-links/shared/metrics"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shortened := getCounterValue(metrics.UrlsShortenedTotal)
	redirected := getCounterValue(metrics.UrlsRedirectedTotal)
	hits := getCounterValue(metrics.CacheHitsTotal)
	misses := getCounterValue(metrics.CacheMissesTotal)

	hitRate := 0.0
	if hits+misses > 0 {
		hitRate = (hits / (hits + misses)) * 100
	}

	reqShorten := getCounterVecValue(metrics.HttpRequestsTotal, prometheus.Labels{"method": "POST", "path": "/shorten", "status": "201"})
	reqRedirect := getCounterVecValue(metrics.HttpRequestsTotal, prometheus.Labels{"method": "GET", "path": "/{shortCode}", "status": "302"})

	resp := StatsResponse{
		UrlsShortened:    shortened,
		UrlsRedirected:   redirected,
		CacheHits:        hits,
		CacheMisses:      misses,
		CacheHitRate:     hitRate,
		RequestsShorten:  reqShorten,
		RequestsRedirect: reqRedirect,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getCounterValue(c prometheus.Counter) float64 {
	m := &dto.Metric{}
	c.Write(m)
	return m.GetCounter().GetValue()
}

func getCounterVecValue(cv *prometheus.CounterVec, labels prometheus.Labels) float64 {
	c, err := cv.GetMetricWith(labels)
	if err != nil {
		return 0
	}
	return getCounterValue(c)
}
