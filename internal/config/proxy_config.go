// Package config charge et valide la configuration immuable du reverse-proxy.
package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// ProxyConfig regroupe les paramètres validés d'exécution de GateShield.
type ProxyConfig struct {
	ListenAddr    string
	UpstreamURL   *url.URL
	RatePerSecond float64
	BurstCapacity int
	MaxBodyBytes  int64
}

var (
	ErrMissingUpstream = errors.New("GATESHIELD_UPSTREAM_URL est requis (ex. http://127.0.0.1:9000)")
	ErrInvalidUpstream = errors.New("GATESHIELD_UPSTREAM_URL doit être une URL HTTP(S) absolue valide")
	ErrInvalidLimits   = errors.New("les seuils de débit et de capacité doivent être strictement positifs")
)

// ParseProxyConfig construit une configuration validée à partir d'une fonction de lecture clé-valeur.
// Fonction pure permettant de tester toutes les règles de validation sans muter l'environnement OS.
func ParseProxyConfig(lookup func(string) string) (ProxyConfig, error) {
	rawUpstream := strings.TrimSpace(lookup("GATESHIELD_UPSTREAM_URL"))
	if rawUpstream == "" {
		return ProxyConfig{}, ErrMissingUpstream
	}

	parsedURL, err := url.Parse(rawUpstream)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return ProxyConfig{}, ErrInvalidUpstream
	}

	listenAddr := strings.TrimSpace(lookup("GATESHIELD_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	ratePerSec := parseFloatOrDefault(lookup("GATESHIELD_RATE_PER_SEC"), 10.0)
	burstCap := parseIntOrDefault(lookup("GATESHIELD_BURST_CAPACITY"), 20)
	maxBody := int64(parseIntOrDefault(lookup("GATESHIELD_MAX_BODY_BYTES"), 1048576))

	if ratePerSec <= 0 || burstCap <= 0 || maxBody <= 0 {
		return ProxyConfig{}, ErrInvalidLimits
	}

	return ProxyConfig{
		ListenAddr:    listenAddr,
		UpstreamURL:   parsedURL,
		RatePerSecond: ratePerSec,
		BurstCapacity: burstCap,
		MaxBodyBytes:  maxBody,
	}, nil
}

func parseFloatOrDefault(raw string, fallback float64) float64 {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return -1
	}
	return val
}

func parseIntOrDefault(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	val, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return -1
	}
	return val
}
