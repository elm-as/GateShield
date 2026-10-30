// Package ratelimit implémente un algorithme Token Bucket pur et un magasin par client borné.
package ratelimit

import (
	"math"
	"sync"
	"time"
)

// BucketPolicy définit la cadence de rechargement et la capacité maximale de rafale.
type BucketPolicy struct {
	RefillPerSecond float64
	BurstCapacity   int
	IdleTTL         time.Duration
}

// BucketState représente l'état immuable d'un seau de jetons à un instant donné.
type BucketState struct {
	AvailableTokens float64
	LastUpdated     time.Time
}

// BucketDecision détaille le résultat pur d'une demande de consommation d'un jeton.
type BucketDecision struct {
	Allowed         bool
	RemainingTokens int
	RetryAfterSec   int
	NextState       BucketState
}

// AdvanceTokenBucket calcule de manière pure le rechargement et la consommation d'un jeton.
// Ne mute aucun argument entrant : retourne le nouvel état et le délai exact de réessai si épuisé.
func AdvanceTokenBucket(current BucketState, now time.Time, policy BucketPolicy) BucketDecision {
	capacity := float64(policy.BurstCapacity)
	tokens := capacity

	if !current.LastUpdated.IsZero() {
		elapsed := math.Max(0, now.Sub(current.LastUpdated).Seconds())
		tokens = math.Min(capacity, current.AvailableTokens+elapsed*policy.RefillPerSecond)
	}

	if tokens >= 1.0 {
		remaining := tokens - 1.0
		return BucketDecision{
			Allowed:         true,
			RemainingTokens: int(math.Floor(remaining)),
			RetryAfterSec:   0,
			NextState:       BucketState{AvailableTokens: remaining, LastUpdated: now},
		}
	}

	deficit := 1.0 - tokens
	waitSec := int(math.Ceil(deficit / math.Max(0.001, policy.RefillPerSecond)))
	return BucketDecision{
		Allowed:         false,
		RemainingTokens: 0,
		RetryAfterSec:   max(1, waitSec),
		NextState:       BucketState{AvailableTokens: tokens, LastUpdated: now},
	}
}

// ClientLimiter coordonne l'état des seaux par clé cliente avec purge automatique des inactifs.
type ClientLimiter struct {
	mu          sync.Mutex
	policy      BucketPolicy
	buckets     map[string]BucketState
	lastSweepAt time.Time
}

// NewClientLimiter instancie un limiteur concurrent protégé contre la fuite mémoire sous flood d'IP.
func NewClientLimiter(policy BucketPolicy) *ClientLimiter {
	if policy.IdleTTL <= 0 {
		policy.IdleTTL = 5 * time.Minute
	}
	return &ClientLimiter{
		policy:  policy,
		buckets: make(map[string]BucketState),
	}
}

// Evaluate applique la politique Token Bucket pour un identifiant client à l'instant donné.
func (l *ClientLimiter) Evaluate(clientKey string, now time.Time) BucketDecision {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastSweepAt.IsZero() || now.Sub(l.lastSweepAt) >= l.policy.IdleTTL {
		l.evictExpiredLocked(now)
		l.lastSweepAt = now
	}

	decision := AdvanceTokenBucket(l.buckets[clientKey], now, l.policy)
	l.buckets[clientKey] = decision.NextState
	return decision
}

func (l *ClientLimiter) evictExpiredLocked(now time.Time) {
	for key, state := range l.buckets {
		if now.Sub(state.LastUpdated) > l.policy.IdleTTL {
			delete(l.buckets, key)
		}
	}
}
