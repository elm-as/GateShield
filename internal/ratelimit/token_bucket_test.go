package ratelimit

import (
	"testing"
	"time"
)

func TestAdvanceTokenBucketBurstAndRefill(t *testing.T) {
	policy := BucketPolicy{RefillPerSecond: 2.0, BurstCapacity: 2}
	t0 := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	d1 := AdvanceTokenBucket(BucketState{}, t0, policy)
	if !d1.Allowed || d1.RemainingTokens != 1 {
		t.Fatalf("1re requête attendue autorisée avec 1 jeton restant, obtenu %+v", d1)
	}

	d2 := AdvanceTokenBucket(d1.NextState, t0, policy)
	if !d2.Allowed || d2.RemainingTokens != 0 {
		t.Fatalf("2e requête attendue autorisée avec 0 jeton restant, obtenu %+v", d2)
	}

	d3 := AdvanceTokenBucket(d2.NextState, t0, policy)
	if d3.Allowed || d3.RetryAfterSec < 1 {
		t.Fatalf("3e requête immédiate attendue bloquée (429), obtenu %+v", d3)
	}

	// Après 600 ms à 2 jetons/s -> +1.2 jetons rechargés : nouvelle requête autorisée
	d4 := AdvanceTokenBucket(d3.NextState, t0.Add(600*time.Millisecond), policy)
	if !d4.Allowed {
		t.Fatalf("requête après rechargement partiel attendue autorisée, obtenu %+v", d4)
	}
}
