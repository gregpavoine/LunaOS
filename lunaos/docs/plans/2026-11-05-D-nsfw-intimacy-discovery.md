# Piste D — NSFW : Intimacy et Discovery — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Verrouiller par des tests les invariants U1 du code Intimacy repris, et livrer `luna-discovery` : synchronisation Reddit et Civitai par API officielles, sur liste blanche, avec stockage local chiffré et reprise fiable.

**Architecture:** `services/discovery` est un package indépendant du vault (identité de service et clé distinctes). Un connecteur par fournisseur implémente une interface commune ; un moteur de jobs gère états, reprise et `Retry-After` ; un téléchargeur borné chiffre les médias et les dédoublonne par SHA-256. La commande `luna-discovery sync` est lancée par un timer systemd et lit une politique propriétaire en JSON.

**Tech Stack:** Go 1.24, SQLite, AES-256-GCM (stdlib), API Reddit OAuth2, API Civitai `GET /api/v1/images`.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md` (§5) ; exigences détaillées : Mont Blanc U1, U10, Annexe B (checkout principal, lecture seule).

## Global Constraints

- Possède `lunaos/services/{intimacy,discovery}/` et les fichiers Core attribués dans `OWNERSHIP.md`.
- API officielles uniquement : pas de scraping, pas de mutualisation de comptes, pas de contournement de restriction. Civitai : jamais de requête sans filtre de liste blanche ; liste vide → aucune requête.
- Toute ressource sans licence ou droits déclarés, marquée mineur/âge incertain, ou non autorisée par le fournisseur, est rejetée avant téléchargement.
- Les actifs du profil corporel et du vault ne sont jamais lus par `discovery` (aucun import de `services/intimacy` depuis `services/discovery`).
- Aucune donnée média, miniature ou métadonnée sensible dans les journaux : seuls codes et identifiants opaques.
- Clé Discovery distincte de la clé du vault : fichier `0600` propre à l'utilisateur `luna-discovery`.
- Ne jamais pousser ; attribution `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- Un élément Civitai dont le niveau de contenu dépasse celui de la liste blanche, ou sans licence → rejeté, compté, jamais téléchargé → test tâche 4.
- Jeton Reddit révoqué en cours de synchronisation → Reddit passe `disconnected`, Civitai et la bibliothèque continuent → test tâche 3.
- Téléchargement coupé, trop gros ou hash différent de l'annonce → rien dans la bibliothèque, reprise possible → test tâche 5.
- Retrait (takedown) d'un élément par le fournisseur → média et métadonnées supprimés à la synchronisation suivante → test tâche 4.
- `429` avec `Retry-After: 120` → aucun appel avant 120 s, puis reprise au curseur sauvegardé → test tâche 3.

---

### Task 1: Invariants U1 sur le code Intimacy repris

**Files:**
- Test: `lunaos/services/intimacy/invariants_test.go`

- [ ] **Step 1:** Lister les invariants U1 déjà couverts par les tests importés (noms de tests dans le message de commit). Écrire seulement ceux qui manquent parmi : `TestOwnerCannotConsentForSubject`, `TestMinorOrUnverifiedGetsNoIntimateScope` (pour `unverified`, `suspected_minor`, `minor`), `TestFlagSuspectedMinorErasesIntimateAssets`, `TestRevocationErasesDependentAssetsInSameTransaction`, `TestSubjectDeletionDestroysKey` (données illisibles avec l'ancienne clé enveloppée), `TestAuditHoldsNoIntimatePayload` (aucun octet du média dans la table d'audit), `TestVoicePhraseAloneIsNotConsent` (sans défi actif → refus).
- [ ] **Step 2:** Lancer dans le conteneur. Un test qui échoue révèle un défaut réel : le corriger dans `services/intimacy/` dans un commit séparé `fix(intimacy): …`, avec le test.
- [ ] **Step 3:** Commit `test(intimacy): pin the Mont Blanc U1 invariants`.

### Task 2: Stockage Discovery chiffré

**Files:**
- Create: `lunaos/services/discovery/store.go`, `store_test.go`, `lunaos/services/discovery/schema.sql`

**Interfaces:**
- Produces: `func OpenStore(dbPath, blobDir, keyPath string) (*Store, error)` (refuse une clé ≠ 32 octets ou de mode plus large que `0600`) ; `type Item struct { Provider, ExternalID, SourceURL, Creator, License, ContentLevel string; Tags []string; SHA256 string; Bytes int64; FetchedAt time.Time }` ; `func (s *Store) Put(ctx context.Context, it Item, media io.Reader) (stored bool, err error)` (`stored=false` si le SHA-256 existe déjà ; `media == nil` → ligne de métadonnées seule, sans blob, dédoublonnée par `(Provider, ExternalID)`) ; `func (s *Store) Delete(ctx context.Context, provider, externalID string) error` (supprime blob et ligne) ; `func (s *Store) Open(ctx context.Context, sha256 string) (io.ReadCloser, error)` ; `func (s *Store) SaveCursor(ctx, provider, scope, cursor string) error` / `Cursor(...) (string, error)`. Blobs : AES-256-GCM, nonce aléatoire par fichier, nom de fichier = HMAC-SHA256(clé, sha256) pour ne pas exposer le hash du média.

- [ ] **Step 1: Tests** : `TestPutEncryptsAtRest` (le contenu clair n'apparaît dans aucun fichier de `blobDir` ni de la base) ; `TestPutDeduplicatesBySHA256` ; `TestPutMetadataOnlyCreatesNoBlob` ; `TestDeleteRemovesBlobAndRow` ; `TestOpenWithWrongKeyFails` ; `TestRejectsWeakKeyFile` ; `TestCursorRoundTrip`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS. **Step 5:** Commit `feat(discovery): add the encrypted local store`.

### Task 3: Interface des connecteurs et moteur de jobs

**Files:**
- Create: `lunaos/services/discovery/connector.go`, `jobs.go`, `jobs_test.go`

**Interfaces:**
- Produces: `type Page struct { Items []Candidate; NextCursor string; Removed []string }` ; `type Candidate struct { Item; MediaURL string; DeclaredSHA256 string; MinorFlag bool }` ; `type Connector interface { Name() string; Fetch(ctx context.Context, scope, cursor string) (Page, error) }` ; erreurs typées `ErrRateLimited{RetryAfter time.Duration}`, `ErrUnauthorized`, `ErrTransient` ; `type Engine struct { Store *Store; Connectors []Connector; Policy Policy; Now func() time.Time; Sleep func(context.Context, time.Duration) error }` ; `func (e *Engine) Sync(ctx context.Context) Report` avec `Report{PerProvider map[string]ProviderReport}` et `ProviderReport{State string /* ok|disconnected|rate_limited|failed */; Stored, Duplicates, Rejected, Removed int}`. Backoff : 1 s, 2 s, 4 s … plafond 5 min, 5 tentatives, puis l'élément va en file des échecs (table `dead_letter`).

- [ ] **Step 1: Tests** (connecteurs de test pilotés par le test ; on teste le moteur) : `TestRetryAfterIsHonoured` (horloge et `Sleep` injectés : aucun appel avant 120 s) ; `TestResumesFromSavedCursor` ; `TestUnauthorizedDisconnectsOnlyThatProvider` ; `TestTransientErrorsBackOffThenDeadLetter` ; `TestRemovedItemsAreDeleted`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS. **Step 5:** Commit `feat(discovery): add the provider-isolated sync engine`.

### Task 4: Politique, porte des droits et connecteur Civitai

**Files:**
- Create: `lunaos/services/discovery/policy.go`, `policy_test.go`, `civitai.go`, `civitai_test.go`

**Interfaces:**
- Produces: `type Policy struct { Civitai CivitaiPolicy; Reddit RedditPolicy }` chargée par `func LoadPolicy(path string) (Policy, error)` (JSON, mode `0600` exigé) ; `CivitaiPolicy{ Tags, Creators, ModelFamilies []string; MaxContentLevel string /* "None"|"Soft"|"Mature"|"X" */; AllowedLicenses []string }` ; `func (p Policy) Admit(c Candidate) (ok bool, reason string)` — refuse si `MinorFlag`, licence absente ou hors `AllowedLicenses`, niveau > `MaxContentLevel`, créateur/tag hors liste. `func NewCivitai(baseURL, apiKey string, p CivitaiPolicy, hc *http.Client) Connector` : une requête `GET /api/v1/images` par entrée de liste blanche (`tags`, `username`), pagination `cursor`/`metadata.nextCursor`, `nsfw` = `MaxContentLevel`, en-tête `Authorization: Bearer <apiKey>`.

- [ ] **Step 1: Tests** (`httptest` imitant la forme documentée de la réponse `items[]`, `metadata.nextCursor`) : `TestEmptyAllowlistMakesNoRequest` ; `TestEachRequestCarriesAllowlistFilter` ; `TestContentLevelAboveMaxRejected` ; `TestMissingLicenseRejected` ; `TestMinorFlagRejected` ; `TestCursorPagination` ; `TestPolicyFileModeEnforced`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter ; `Engine.Sync` appelle `Policy.Admit` avant tout téléchargement. **Step 4:** PASS.
- [ ] **Step 5:** Vérifier les noms de champs contre la documentation officielle (`civitai-developer-docs/site/reference/images.md`) et les citer dans le commit ; tout écart avec le test est corrigé dans le test et le code.
- [ ] **Step 6:** Commit `feat(discovery): add the allowlisted Civitai connector and rights gate`.

### Task 5: Téléchargeur borné

**Files:**
- Create: `lunaos/services/discovery/download.go`, `download_test.go`

**Interfaces:**
- Produces: `func Download(ctx context.Context, hc *http.Client, c Candidate, maxBytes int64, tmpDir string) (path string, sha string, err error)` — `.part` dans `tmpDir`, reprise par `Range` si le serveur l'accepte, arrêt au-delà de `maxBytes` (défaut 512 MiB), hash calculé ; si `DeclaredSHA256` est fourni et diffère → `ErrHashMismatch` et fichier supprimé. L'`Engine` passe le fichier à `Store.Put` puis le supprime.

- [ ] **Step 1: Tests** : `TestResumeWithRange` ; `TestOversizeAborted` ; `TestHashMismatchLeavesNothing` ; `TestCancelledContextLeavesNoFinalFile`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS. **Step 5:** Commit `feat(discovery): add the bounded resumable downloader`.

### Task 6: Connecteur Reddit

**Files:**
- Create: `lunaos/services/discovery/reddit.go`, `reddit_test.go`

**Interfaces:**
- Décision : les publications Reddit ne déclarent pas de licence ; pour respecter U1.13, Reddit alimente uniquement l'index en **métadonnées et liens** (titre, subreddit, auteur, permalien, URL média) — `Candidate.MediaURL` vide, aucun téléchargement, `Store.Put` appelé avec `media == nil`.
- Produces: `RedditPolicy{ Subreddits []string }` ; `func NewReddit(cfg RedditConfig, p RedditPolicy, hc *http.Client) Connector` avec `RedditConfig{ ClientID, ClientSecret, RefreshToken, UserAgent string; TokenURL, APIBase string }` (défauts `https://www.reddit.com/api/v1/access_token`, `https://oauth.reddit.com`). Rafraîchit le jeton (`grant_type=refresh_token`), lit `/r/<sub>/new` avec `after`, respecte `X-Ratelimit-Remaining`/`X-Ratelimit-Reset` (→ `ErrRateLimited`), `401` après rafraîchissement → `ErrUnauthorized`. Ne propose que les liens média directs de publications de subreddits listés ; aucune découverte automatique de nouveaux subreddits (une suggestion éventuelle relève d'une validation du propriétaire, hors de ce plan).

- [ ] **Step 1: Tests** : `TestRedditCandidatesAreMetadataOnly` (aucune requête vers l'URL média, aucun blob créé) ; `TestRefreshesTokenOnce` ; `TestOnlyListedSubredditsRequested` ; `TestRateLimitHeadersBecomeRetryAfter` ; `TestUnauthorizedAfterRefreshDisconnects` ; `TestUserAgentRequired` (vide → erreur de configuration, aucune requête).
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS.
- [ ] **Step 5:** Vérifier contre la Data API Wiki de Reddit et la Responsible Builder Policy (accès approuvé requis) ; le noter dans le commit.
- [ ] **Step 6:** Commit `feat(discovery): add the OAuth Reddit connector`.

### Task 7: Commande et unités

**Files:**
- Create: `lunaos/services/discovery/cmd/luna-discovery/main.go`, `main_test.go`, `lunaos/services/discovery/systemd/luna-discovery.service`, `luna-discovery.timer`, `luna-discovery.sysusers`

**Interfaces:**
- Produces: `luna-discovery sync --policy /etc/lunaos/discovery/policy.json --secrets /etc/lunaos/discovery/secrets.json --db /var/lib/luna-discovery/index.db --blobs /var/lib/luna-discovery/blobs --key /var/lib/luna-discovery/key` ; imprime le `Report` en JSON (codes et compteurs seulement) ; code 0 si au moins un fournisseur `ok` ou aucun configuré, 1 sinon. `luna-discovery keygen --key <path>` crée la clé `0600` sans l'afficher, refuse d'écraser.

- [ ] **Step 1: Tests** : `TestSyncWithNoProvidersConfiguredIsNoop` ; `TestKeygenRefusesOverwrite` ; `TestReportContainsNoMediaURLs`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS.
- [ ] **Step 5:** Unités : utilisateur `luna-discovery`, `StateDirectory=luna-discovery`, `ProtectSystem=strict`, `ProtectHome=yes`, `NoNewPrivileges=yes`, `PrivateTmp=yes` ; timer `OnCalendar=hourly`, `RandomizedDelaySec=10m`, `Persistent=true`. `systemd-analyze verify` dans `archlinux:latest` → aucune erreur.
- [ ] **Step 6:** Commit `feat(discovery): add the luna-discovery command and timer`.
