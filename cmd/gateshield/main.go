package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elm-as/GateShield/internal/config"
	"github.com/elm-as/GateShield/internal/proxy"
	"github.com/elm-as/GateShield/internal/ratelimit"
)

func main() {
	cfg, err := config.ParseProxyConfig(os.Getenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERREUR CONFIG] %v\n", err)
		os.Exit(2)
	}

	limiter := ratelimit.NewClientLimiter(ratelimit.BucketPolicy{
		RefillPerSecond: cfg.RatePerSecond,
		BurstCapacity:   cfg.BurstCapacity,
		IdleTTL:         5 * time.Minute,
	})
	gateway := proxy.NewShieldGateway(cfg, limiter)

	// Délais HTTP stricts pour neutraliser l'épuisement de descripteurs par attaques lentes (Slowloris).
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           gateway,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		fmt.Printf(
			"GateShield actif sur %s -> amont %s (rate=%.1f req/s, burst=%d)\n",
			cfg.ListenAddr, cfg.UpstreamURL.String(), cfg.RatePerSecond, cfg.BurstCapacity,
		)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "[ERREUR SERVEUR] %v\n", serveErr)
			os.Exit(1)
		}
	}()

	waitForShutdownSignal(server)
}

func waitForShutdownSignal(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	fmt.Println("GateShield arrêté proprement.")
}
