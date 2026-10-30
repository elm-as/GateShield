package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/elm-as/GateShield/internal/config"
	"github.com/elm-as/GateShield/internal/ratelimit"
	"github.com/elm-as/GateShield/internal/telemetry"
)

func TestShieldGatewayEnforcesRateLimitAndExposesMetrics(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	parsedURL, _ := url.Parse(upstream.URL)
	cfg := config.ProxyConfig{
		ListenAddr:    ":0",
		UpstreamURL:   parsedURL,
		RatePerSecond: 1.0,
		BurstCapacity: 2,
		MaxBodyBytes:  4096,
	}
	limiter := ratelimit.NewClientLimiter(ratelimit.BucketPolicy{
		RefillPerSecond: 1.0,
		BurstCapacity:   2,
		IdleTTL:         time.Minute,
	})
	gateway := NewShieldGateway(cfg, limiter)

	for i := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		req.RemoteAddr = "198.51.100.42:54321"
		rec := httptest.NewRecorder()
		gateway.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("requête #%d attendue 200 OK, obtenu %d", i+1, rec.Code)
		}
	}

	blockedReq := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	blockedReq.RemoteAddr = "198.51.100.42:54321"
	blockedRec := httptest.NewRecorder()
	gateway.ServeHTTP(blockedRec, blockedReq)
	if blockedRec.Code != http.StatusTooManyRequests {
		t.Fatalf("3e requête en rafale attendue 429, obtenu %d", blockedRec.Code)
	}
	if blockedRec.Header().Get("Retry-After") == "" {
		t.Fatalf("l'en-tête standard Retry-After doit être renseigné sur un 429")
	}

	metricsReq := httptest.NewRequest(http.MethodGet, metricsEndpointPath, nil)
	metricsRec := httptest.NewRecorder()
	gateway.ServeHTTP(metricsRec, metricsReq)

	var snap telemetry.MetricsSnapshot
	if err := json.NewDecoder(metricsRec.Body).Decode(&snap); err != nil {
		t.Fatalf("JSON de télémétrie invalide : %v", err)
	}
	if snap.TotalRequests != 3 || snap.AllowedRequests != 2 || snap.RejectedRateLimit != 1 {
		t.Fatalf("compteurs de télémétrie incohérents : %+v", snap)
	}
}
