package stats

type StatsResponse struct {
	UrlsShortened    float64 `json:"urls_shortened"`
	UrlsRedirected   float64 `json:"urls_redirected"`
	CacheHits        float64 `json:"cache_hits"`
	CacheMisses      float64 `json:"cache_misses"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	RequestsShorten  float64 `json:"requests_shorten"`
	RequestsRedirect float64 `json:"requests_redirect"`
}
