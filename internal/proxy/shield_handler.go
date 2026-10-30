// Package proxy orchestre le filtrage défensif, la limitation de débit et le relais HTTP.
package proxy

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/elm-as/GateShield/internal/config"
	"github.com/elm-as/GateShield/internal/ratelimit"
	"github.com/elm-as/GateShield/internal/telemetry"
)

const metricsEndpointPath = "/_gateshield/metrics"

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// ShieldGateway encapsule le pipeline de protection placé devant le serveur amont.
type ShieldGateway struct {
	cfg       config.ProxyConfig
	limiter   *ratelimit.ClientLimiter
	collector *telemetry.TrafficCollector
	reverse   *httputil.ReverseProxy
	clock     func() time.Time
}

// NewShieldGateway assemble une passerelle reverse-proxy prête à servir le trafic entrant.
func NewShieldGateway(cfg config.ProxyConfig, limiter *ratelimit.ClientLimiter) *ShieldGateway {
	rp := httputil.NewSingleHostReverseProxy(cfg.UpstreamURL)
	collector := &telemetry.TrafficCollector{}

	rp.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream_unreachable"}`))
	}

	return &ShieldGateway{
		cfg:       cfg,
		limiter:   limiter,
		collector: collector,
		reverse:   rp,
		clock:     time.Now,
	}
}

// ServeHTTP applique les garde-fous de sécurité avant tout transfert vers le backend cible.
func (g *ShieldGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	applySecurityHeaders(w.Header())

	if r.Method == http.MethodGet && r.URL.Path == metricsEndpointPath {
		g.serveMetricsJSON(w)
		return
	}

	clientIP := extractRemoteHost(r.RemoteAddr)
	decision := g.limiter.Evaluate(clientIP, g.clock())
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.RemainingTokens))

	if !decision.Allowed {
		g.collector.RecordRateLimited()
		w.Header().Set("Retry-After", strconv.Itoa(decision.RetryAfterSec))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded"}`))
		return
	}

	// Borne stricte du corps entrant pour prévenir l'épuisement RAM par requêtes géantes.
	r.Body = http.MaxBytesReader(w, r.Body, g.cfg.MaxBodyBytes)
	recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	started := g.clock()

	g.reverse.ServeHTTP(recorder, r)
	g.collector.RecordAllowed(recorder.statusCode, g.clock().Sub(started))
}

func (g *ShieldGateway) serveMetricsJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(g.collector.Snapshot())
}

// extractRemoteHost lit uniquement le socket pair direct (RemoteAddr) par défaut afin
// d'empêcher le contournement du rate-limiter par falsification d'en-tête X-Forwarded-For.
func extractRemoteHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil || host == "" {
		return remoteAddr
	}
	return host
}

func applySecurityHeaders(headers http.Header) {
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("Referrer-Policy", "no-referrer")
}
