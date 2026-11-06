# Piste C — Moteur LLM local — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Un service `luna-inference` qui supervise llama.cpp `llama-server`, n'utilise que des modèles vérifiés par hash, et n'est joignable que par le Core.

**Architecture:** `luna-inference` lance `llama-server` en loopback sur un port choisi par lui, attend sa santé au démarrage, le redémarre avec un budget, et expose un socket Unix réservé à l'UID du Core. Le modèle est chargé à la demande et peut être déchargé (priorité VRAM à ComfyUI). La réponse porte toujours `untrusted: true`.

**Tech Stack:** Go 1.24 (stdlib uniquement), llama.cpp `llama-server`, JSON, systemd.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md` (§4)

## Global Constraints

- Possède `lunaos/services/inference/` et `lunaos/models/`.
- Manifeste en JSON (`lunaos/models/manifest.json`), pas TOML : la stdlib Go n'a pas de TOML et on n'ajoute pas de dépendance.
- Aucun fichier de modèle dans le dépôt. Un fichier dont le sha256 ne correspond pas est refusé et supprimé.
- `llama-server` écoute uniquement sur `127.0.0.1` ; seul `luna-inference` s'y connecte.
- Socket `/run/luna-inference/inference.sock`, mode `0660`, groupe `luna-inference` ; seul l'UID `lunaos` (Core) est accepté (`internal/peer`).
- Pas de repli cloud. Pas de sortie réseau autre que le téléchargement explicite d'un modèle du manifeste.
- `max_tokens` ∈ [1, 1024] ; prompt ≤ 32 KiB.
- Ne jamais pousser ; attribution `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- `llama-server` absent ou plante en boucle → le service répond `INFERENCE_DEGRADED` (HTTP 503) sans bloquer, et `/v1/health` le dit → test tâche 2.
- Téléchargement interrompu ou hash faux → aucun fichier à l'emplacement final → test tâche 1.
- Requête pendant le démarrage du modèle → attend jusqu'au délai (60 s par défaut), puis 503, jamais de réponse vide présentée comme un succès → test tâche 3.
- Un appelant qui n'est pas le Core sur le socket → 403 `SOCKET_PEER_DENIED` → test tâche 3.
- Déchargement demandé pendant une génération → la génération en cours se termine ou échoue proprement, puis le modèle est libéré → test tâche 3.

---

### Task 1: Manifeste et vérification des modèles

**Files:**
- Create: `lunaos/services/inference/manifest.go`, `manifest_test.go`, `lunaos/models/manifest.json`

**Interfaces:**
- Produces: `type Model struct { ID, URL, SHA256, License string; SizeBytes int64; VRAMMiB int }` ; `func LoadManifest(path string) ([]Model, error)` (refuse les ID dupliqués, sha256 non hexadécimal de 64 caractères, URL non `https`, licence vide) ; `func Fetch(ctx context.Context, m Model, dir string, client *http.Client) (string, error)` — télécharge vers `dir/.<id>.part`, refuse au-delà de `SizeBytes`, vérifie le sha256, renomme en `dir/<id>.gguf` ; si le fichier final existe et a le bon hash, ne télécharge pas.

- [ ] **Step 1: Tests** (serveur `httptest.NewTLSServer`) : `TestLoadManifestRejectsInvalidEntries` (un cas par règle) ; `TestFetchVerifiesHash` ; `TestFetchHashMismatchLeavesNoFile` (ni `.part` ni `.gguf`) ; `TestFetchOversizeAborted` ; `TestFetchSkipsVerifiedFile` (aucune requête HTTP) ; `TestFetchInterruptedLeavesNoFinalFile`.
- [ ] **Step 2:** Lancer → FAIL. **Step 3:** Implémenter. **Step 4:** Relancer → PASS.
- [ ] **Step 5:** `manifest.json` : une entrée de test `qwen2.5-0.5b-instruct-q4_k_m` (URL `https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf`, licence `Apache-2.0`, `VRAMMiB: 1024`) ; le sha256 et la taille sont relevés à partir du fichier téléchargé une fois et de la page du dépôt, et les deux doivent concorder — sinon, ne pas committer l'entrée et le signaler.
- [ ] **Step 6:** Commit `feat(inference): add the verified model manifest`.

### Task 2: Superviseur de `llama-server`

**Files:**
- Create: `lunaos/services/inference/supervisor.go`, `supervisor_test.go`

**Interfaces:**
- Produces: `type Supervisor struct { Binary, ModelPath string; GPULayers int; StartTimeout time.Duration }` ; `func (s *Supervisor) Start(ctx context.Context) (baseURL string, err error)` (choisit un port libre, lance `Binary -m ModelPath --host 127.0.0.1 --port <p> -ngl <GPULayers>`, attend `GET /health` = 200 jusqu'à `StartTimeout`) ; `func (s *Supervisor) Stop() error` (SIGTERM, puis SIGKILL après 10 s) ; `func (s *Supervisor) State() State` avec `State` ∈ {`StateStopped`, `StateStarting`, `StateReady`, `StateDegraded`} ; budget de redémarrage : 3 échecs en 60 s → `StateDegraded`, plus aucune tentative jusqu'à `Reset()`.

- [ ] **Step 1: Tests** avec un faux binaire écrit par le test (petit programme Go compilé dans `t.TempDir()` qui sert `/health`, ou qui quitte immédiatement selon un argument) — le comportement testé est celui du superviseur, pas du faux : `TestStartWaitsForHealth`, `TestStartTimesOut`, `TestCrashLoopBecomesDegraded`, `TestStopKillsAfterGrace`, `TestMissingBinaryIsDegraded`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter (`os/exec`, `net.Listen("tcp","127.0.0.1:0")` pour réserver puis libérer un port). **Step 4:** PASS.
- [ ] **Step 5:** Commit `feat(inference): supervise llama-server with a restart budget`.

### Task 3: Démon `luna-inference`

**Files:**
- Create: `lunaos/services/inference/server.go`, `server_test.go`, `lunaos/services/inference/client.go`, `lunaos/services/inference/cmd/luna-inference/main.go`, `lunaos/services/inference/real_llama_test.go`

**Interfaces:**
- Consumes: `peer.ConnContext`, `peer.FromContext` (phase 0).
- Produces (utilisé par l'intégration Z5) :
  - HTTP sur socket Unix : `GET /v1/health` → `{"state":"stopped|starting|ready|degraded","model":"<id>"}` ; `POST /v1/generate` `{"prompt":string,"max_tokens":int}` → `{"text":string,"model":string,"untrusted":true}` (relaie vers `POST /completion` de `llama-server`, champs `prompt`, `n_predict`) ; `POST /v1/unload` → 204.
  - Codes d'erreur JSON `{"code":...}` : `SOCKET_PEER_DENIED` (403), `INVALID_REQUEST` (400), `INFERENCE_DEGRADED` (503), `INFERENCE_TIMEOUT` (503).
  - Client Go : `func NewClient(socket string) *Client` ; `func (c *Client) Generate(ctx context.Context, prompt string, maxTokens int) (Result, error)` avec `Result{Text, Model string; Untrusted bool}`.
  - Sans `--model`, ou si le fichier vérifié du modèle est absent, le service démarre quand même : `/v1/health` → `stopped`, `/v1/generate` → 503 `INFERENCE_NO_MODEL` (test `TestNoModelConfiguredReturns503`).
  - Binaire : `luna-inference serve --socket <path> --model-dir <dir> [--model <id>] --manifest <path> --llama-server <path> --allowed-uid <uid> [--gpu-layers N]` et `luna-inference fetch --manifest <path> --model-dir <dir> <id>` (appelle `Fetch` ; seule commande qui sort sur le réseau).

- [ ] **Step 1: Tests** (`httptest` + `Supervisor` branché sur le faux binaire de la tâche 2) : `TestPeerOtherThanCoreDenied` ; `TestGenerateValidatesLimits` (0 et 1025 tokens, prompt 32 KiB+1) ; `TestGenerateMarksOutputUntrusted` ; `TestGenerateStartsModelLazily` ; `TestGenerateWhileDegradedReturns503` ; `TestUnloadDuringGenerateCompletesThenStops` ; `TestClientRoundTrip`.
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS.
- [ ] **Step 5: Test réel** `real_llama_test.go` : ignoré sauf si `LUNA_TEST_LLAMA_SERVER` et `LUNA_TEST_MODEL` sont définis ; démarre le vrai `llama-server` avec le modèle du manifeste, `Generate("Say OK", 8)` → texte non vide, `untrusted=true`. L'exécuter une fois dans un conteneur Arch avec llama.cpp (paquet officiel ou AUR, nom noté dans le commit) ; si aucun paquet n'existe pour la plateforme, le dire et ne pas déclarer le test réel passé.
- [ ] **Step 6:** Commit `feat(inference): serve generation to the Core over a peer-checked socket`.

### Task 4: Unité systemd et utilisateur

**Files:**
- Create: `lunaos/services/inference/systemd/luna-inference.service`, `lunaos/services/inference/systemd/luna-inference.sysusers`

- [ ] **Step 1:** Unité : `User=luna-inference`, `SupplementaryGroups=video render`, `RuntimeDirectory=luna-inference`, `StateDirectory=luna-inference` (modèles dans `/var/lib/luna-inference/models`), `NoNewPrivileges=yes`, `ProtectSystem=strict`, `ProtectHome=yes`, `PrivateTmp=yes`, `IPAddressDeny=any`, `IPAddressAllow=localhost`, `DeviceAllow=/dev/nvidia* rw`, `DeviceAllow=/dev/dri rw`. Le téléchargement de modèle n'est pas fait par le service : commande séparée `luna-inference fetch <id>` lancée par le propriétaire.
- [ ] **Step 2:** `systemd-analyze verify` dans `archlinux:latest` → aucune erreur (même méthode que la piste B, tâche 2).
- [ ] **Step 3:** Commit `feat(inference): add the luna-inference unit and user`.
