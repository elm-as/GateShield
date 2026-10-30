# ⚡ GateShield

> Reverse-proxy HTTP défensif avec *rate-limiting* Token Bucket, garde-fous anti-Slowloris et télémétrie temps réel (*From Data to Systems* — ElmasCore).

GateShield se place en frontal d'un service applicatif pour absorber les rafales abusives par IP, borner la taille des charges utiles entrantes, injecter des en-têtes de durcissement HTTP et exposer des métriques de trafic sans verrou bloquant.

---

## 1. Choix technique et architecture

### Pourquoi Go (`net/http/httputil` standard) ?
Un reverse-proxy est un composant critique lié aux entrées/sorties réseau : le runtime **Go** fournit nativement le multiplexage HTTP/1.1 & HTTP/2, le contrôle fin des délais de lecture d'en-têtes (`ReadHeaderTimeout`), l'isolation par goroutine légère et les opérations atomiques 64 bits (`sync/atomic`) sans dépendance externe susceptible d'introduire une surface d'attaque supplémentaire.

### Organisation par domaine métier (`ELMAS.md`)

```text
GateShield/
├── cmd/gateshield/           # Démarrage serveur durci & arrêt gracieux SIGINT/SIGTERM
└── internal/
    ├── config/               # Lecture et validation pure des variables d'environnement
    ├── ratelimit/            # Transition pure Token Bucket + magasin concurrent à purge TTL
    ├── telemetry/            # Compteurs atomiques sans verrou (2xx/5xx, rejets 429, latence)
    └── proxy/                # Pipeline HTTP : en-têtes sécurité, MaxBytesReader, ReverseProxy
```

---

## 2. Principes de sécurité appliqués

- **Non-confiance envers `X-Forwarded-For` par défaut** : l'identifiant client est extrait de `RemoteAddr` (socket TCP réel) pour empêcher le contournement du limiteur par usurpation d'en-tête HTTP.
- **Purge TTL des seaux inactifs** : évite l'épuisement mémoire (OOM) lors d'un flood distribué générant des milliers d'adresses sources éphémères.
- **Protection anti-Slowloris & anti-OOM** : `ReadHeaderTimeout: 5s` sur le socket d'écoute et `http.MaxBytesReader` sur chaque corps de requête.
- **Observabilité native** : `GET /_gateshield/metrics` retourne instantanément l'état JSON (`total_requests`, `allowed_requests`, `rejected_429`, `rejection_ratio_pct`, `mean_latency_ms`).

---

## 3. Lancement rapide

### Configuration (variables d'environnement)
Copier `.env.example` ou exporter directement :

```bash
export GATESHIELD_LISTEN_ADDR=":8080"
export GATESHIELD_UPSTREAM_URL="http://127.0.0.1:9000"
export GATESHIELD_RATE_PER_SEC="10"
export GATESHIELD_BURST_CAPACITY="20"
go run ./cmd/gateshield
```

### Exécuter les tests unitaires et d'intégration
```bash
go test -v ./...
```

---

## 4. Feuille de route
- [x] Moteur Token Bucket pur avec calcul exact de `Retry-After`
- [x] Purge mémoire automatique des clients expirés (TTL)
- [x] Télémétrie atomique JSON (`/_gateshield/metrics`)
- [ ] Magasin distribué optionnel sous Redis (mode multi-instances)
- [ ] Réécriture comparative du moteur critique en Rust (`tokio` / `hyper`)
