# LunaOS fork — Phase 0 (fondations partagées) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Préparer la branche d'intégration pour que les quatre pistes A/B/C/D puissent avancer en parallèle sans conflit : branche `upstream`, code Go repris après revue, contrats partagés, règles de propriété.

**Architecture:** La branche d'intégration est `claude/omarchy-fork-specs-ad9dd4` (dite « intégration »). `upstream` fige Omarchy `99f479b6`. Le code Go existant de `feature/lunaos-mac-image` (`7ec50f4a`) est importé une seule fois ici, car le Core dépend déjà du package `intimacy`. Les pistes partent du commit final de cette phase.

**Tech Stack:** git, Go 1.24 (conteneur `golang:1.24`), SQLite (`github.com/mattn/go-sqlite3`), JSON.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md`

## Global Constraints

- Ne jamais pousser ; commits atomiques, message succinct, terminé par `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.
- Module Go unique : `lunaos.local/lunaos`, racine `lunaos/`.
- Vérification Go uniquement dans un conteneur Linux : `docker run --rm -v "$PWD/lunaos":/src -w /src golang:1.24 make test check`.
- Logs de tests écrits dans un fichier puis lus (jamais `| tail`).
- Rien n'est importé sans relecture ; un fichier sans test qui le couvre est laissé de côté.
- Non repris : `lunaos/rust/`, `lunaos/luna/` (Python), `lunaos/shell/shell.qml`, `lunaos/packaging/arch/`, `lunaos/scripts/*iso*`, `lunaos/tests/install*`, `lunaos/tests/test_slice.py`.

## Review Focus

- Un secret (jeton, clé privée, `.lunaos-signing/`) embarqué par l'import → l'import doit le refuser ; contrôle `git grep` dans la tâche 2.
- Les tests Go dépendant de macOS (`*_other.go`) passent sous Linux mais pas l'inverse → la vérification se fait sous Linux uniquement.
- Le renommage du module casse un import oublié → `go vet ./...` doit passer.
- Les contrats JSON divergent des constantes Go existantes (noms de scopes) → test de cohérence dans la tâche 3.
- Une piste modifie un fichier qu'elle ne possède pas → la table de propriété (tâche 4) est la référence en revue.

---

### Task 1: Branche `upstream`

**Files:** aucun.

- [ ] **Step 1:** `git branch upstream 99f479b6` puis `git rev-parse upstream` → `99f479b6…`.
- [ ] **Step 2:** `git log --oneline upstream..HEAD` → affiche uniquement les commits LunaOS (design, plans).

### Task 2: Import du code Go relu

**Files:**
- Create (depuis `feature/lunaos-mac-image`, via `git checkout 7ec50f4a -- <path>`) : `lunaos/services/{core,policy,identity,intimacy}/`, `lunaos/systemd/`, `lunaos/systemd-sysusers/lunaos.conf`, `lunaos/contracts/openapi/`, `lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile`, `lunaos/scripts/dev-native.sh`, `lunaos/.gitignore` s'il existe.
- Create: `lunaos/services/internal/peer/peer_linux.go`, `peer_other.go`, `peer_test.go`
- Modify: `lunaos/services/policy/main.go`, `peer_linux.go`, `peer_other.go` (utilisent `internal/peer`)

**Interfaces:**
- Produces: package `lunaos.local/lunaos/services/internal/peer` — `func UID(conn net.Conn) (int, error)` (SO_PEERCRED sous Linux ; `ErrUnsupported` ailleurs), `func ConnContext(ctx context.Context, c net.Conn) context.Context` et `func FromContext(ctx context.Context) (int, bool)`. Réutilisé par C (inférence) et D (discovery).

- [ ] **Step 1:** Checkout des chemins listés depuis `7ec50f4a`.
- [ ] **Step 2:** Remplacer `lunaos.local/bootstrap` par `lunaos.local/lunaos` dans `go.mod` et tous les imports.
- [ ] **Step 3:** Réduire `lunaos/Makefile` aux cibles `test`, `check`, `test-process-linux`, `build` ; `test` = `go test -buildvcs=false -count=1 ./services/...` ; `check` = `gofmt -l` vide + `go vet ./services/...` + `bash -n scripts/dev-native.sh` + `python3 -m json.tool` sur chaque JSON de `contracts/`.
- [ ] **Step 4:** Revue écrite fichier par fichier dans le message de commit : ce qui est gardé, ce qui est retiré et pourquoi. Contrôle secrets : `git grep -nE 'BEGIN (OPENPGP|PRIVATE)|Bearer [0-9a-f]{16}' -- lunaos/` → aucun résultat.
- [ ] **Step 5:** Écrire `peer_test.go` : `TestUIDOfUnixSocketPairIsCurrentUser` (sous Linux, `net.Listen("unix", <tmp>/s.sock)` + `net.Dial` : `UID` de la connexion acceptée == `os.Getuid()`), puis extraire la logique de `policy/peer_linux.go` dans `internal/peer` ; `policy` l'utilise.
- [ ] **Step 6:** `docker run --rm -v "$PWD/lunaos":/src -w /src golang:1.24 make test check >/tmp/p0.log 2>&1` → code 0 (lire le log). Tout échec hérité est noté, pas masqué.
- [ ] **Step 7:** Commit `feat(lunaos): import the reviewed Go control plane`.

### Task 3: Contrats partagés

**Files:**
- Create: `lunaos/contracts/actions.json`, `lunaos/contracts/consent-scopes.json`, `lunaos/contracts/contracts.go`, `lunaos/contracts/contracts_test.go`

**Interfaces:**
- Produces: `actions.json` = `{"actions":[{"name":"local.checksum","privacy_class":"general"},{"name":"inference.generate","privacy_class":"general"}]}`. `consent-scopes.json` = les 13 scopes U1.5 (`FACE_REFERENCE`, `FACE_RECOGNITION`, `BODY_PROFILE`, `INTIMATE_VIEW`, `INTIMATE_STILL_CAPTURE`, `INTIMATE_VIDEO_CAPTURE`, `INTIMATE_DATA_RETENTION`, `TRAINING_FACE`, `TRAINING_BODY`, `TRAINING_BLIND`, `ONGOING_AUTONOMOUS_CAPTURE`, `NAS_ENCRYPTED_BACKUP`, `GENERATION_REFERENCE`). Package `lunaos.local/lunaos/contracts` avec `//go:embed` : `func Actions() []Action` (`Action{Name, PrivacyClass string}`), `func ConsentScopes() []string`.

- [ ] **Step 1:** Test `TestConsentScopesMatchVault` : l'ensemble `ConsentScopes()` est égal à l'ensemble des constantes `intimacy.SubjectScope` (`SubjectFaceReference` … `SubjectGenerationReference`, mêmes chaînes `FACE_REFERENCE` …) ; si le vault n'exporte pas sa table `knownScopes`, exporter `func intimacy.KnownScopes() []SubjectScope` (commit séparé, piste D prévenue).
- [ ] **Step 2:** Test `TestActionsAreUniqueAndKnownPrivacy` : noms uniques, `privacy_class` ∈ {`general`}.
- [ ] **Step 3:** Lancer → FAIL (package absent). Implémenter. Relancer dans le conteneur → PASS.
- [ ] **Step 4:** Commit `feat(contracts): add the shared action and consent-scope registry`.

### Task 4: Propriété des fichiers et lancement des pistes

**Files:**
- Create: `lunaos/docs/plans/OWNERSHIP.md`

- [ ] **Step 1:** Recopier la table de propriété du design, en ajoutant : D possède aussi `lunaos/services/core/{subjects_http,library_http,intimacy_context}.go` et leurs tests (`subjects_http_test`, `library_http_test`, `intimacy_http_test`, `preference_http_test`, `discovery_policy_http_test`) ; B possède le reste de `lunaos/services/core/` ; `lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile` et `lunaos/contracts/` ne changent que par commit isolé, annoncé aux autres pistes ; `lunaos/services/internal/` appartient à la phase 0 (changement = commit isolé).
- [ ] **Step 2:** Commit `docs(plans): record track file ownership`.
- [ ] **Step 3:** Démarrer une session par piste (A, B, C, D), chacune dans son worktree à partir de ce commit, avec son plan `2026-11-05-<A|B|C|D>-*.md`.
