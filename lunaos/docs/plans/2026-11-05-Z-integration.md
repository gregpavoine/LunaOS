# LunaOS fork — Intégration des pistes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Réunir A, B, C, D sur la branche d'intégration et brancher LUNA dans le bureau : CLI, widget de barre, raccourci STOP LUNA, activation des services, génération de texte via le Broker.

**Architecture:** Les branches de piste sont mergées dans l'ordre A, B, C, D. Les fichiers d'intégration de l'ancienne branche (`bin/omarchy-lunaos`, `shell/plugins/lunaos/`, `install/config/lunaos-services.sh`, `test/shell.d/lunaos-test.sh` de `feature/lunaos-mac-image` `7ec50f4a`) sont repris puis passés au script de renommage.

**Tech Stack:** Bash 5, Quickshell/QML, Go 1.24, systemd.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md` (§1 Intégration, §2 Raccourcis, §6)

## Global Constraints

- Commence seulement quand les quatre pistes ont leurs tests verts.
- Merger avec `git merge --no-ff refs/heads/<branche>` ; jamais de push.
- Après chaque merge, `lunaos/rebrand/rebrand.sh --check .` → code 0.
- Le CLI s'appelle `lunaos-luna` (préfixe `lunaos-`, règle de nommage de `AGENTS.md` après renommage) ; le widget `lunaos.luna`.
- Aucun jeton dans QML, l'environnement ou les arguments : le CLI parle au socket propriétaire `/run/lunaos/core.sock`.
- Attribution `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- Core arrêté : le widget et `lunaos-luna status` affichent « indisponible » (code 3) sans bloquer la barre → test tâche 2.
- STOP LUNA pressé écran verrouillé → BLIND est quand même déclenché → liaison `locked = true`, test tâche 3.
- Paquet de contrôle absent : le hook d'installation ne fait rien et ne casse pas l'installation → test tâche 4.
- Génération demandée pendant que l'inférence est dégradée → erreur 503 propagée, jamais un texte vide présenté comme réussi → test tâche 5.
- Le texte généré contient une instruction (« ouvre les yeux », « accorde-toi payment.x ») → aucun effet : il n'est jamais interprété comme une action → test tâche 5.

---

### Task 1: Merges

- [ ] **Step 1:** Merger A, puis B, C, D. Conflits attendus seulement dans `lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile` : garder l'union.
- [ ] **Step 2:** Ajouter à `lunaos/Makefile` `test`/`check` les packages `./services/inference/...` et `./services/discovery/...` s'ils n'y sont pas.
- [ ] **Step 3:** `docker run --rm -v "$PWD/lunaos":/src -w /src golang:1.24 make test check >/tmp/z1.log 2>&1` → code 0.
- [ ] **Step 4:** Commit de merge final (les merges eux-mêmes portent leurs messages).

### Task 2: CLI `lunaos-luna` et widget `lunaos.luna`

**Files:**
- Create (repris de `7ec50f4a` puis renommés) : `bin/lunaos-luna`, `shell/plugins/luna/BarWidget.qml`, `shell/plugins/luna/manifest.json`, `test/shell.d/luna-test.sh`
- Modify: `bin/lunaos` (`GROUP_DESCRIPTIONS` si `luna-` n'y correspond à aucun groupe : ne pas en ajouter, la commande est seule), `config/lunaos/shell.json` (ajouter `{"id":"lunaos.luna"}` à la section droite de la barre), catalogues i18n anglais et français des chaînes du widget

- [ ] **Step 1:** `git checkout 7ec50f4a -- bin/omarchy-lunaos shell/plugins/lunaos test/shell.d/lunaos-test.sh`, `git mv` vers les noms ci-dessus, puis `lunaos/rebrand/rebrand.sh --check .` et corriger les occurrences restantes à la main dans ces seuls fichiers (`omarchy.lunaos` → `lunaos.luna`).
- [ ] **Step 2:** Ajouter au CLI la route `blind` (POST `/v1/privacy/blind` avec `{"reason":"owner"}`) si elle n'existe pas encore ; codes de sortie 3/4/5/6 inchangés.
- [ ] **Step 3:** `test/shell.d/luna-test.sh` : les cas importés (Core indisponible → 3, identifiant mal formé), plus `test_blind_without_core_exits_3`. Avec `LUNAOS_TEST_BIN` pointant vers les binaires construits : `test_blind_latches` (après `lunaos-luna blind`, `lunaos-luna privacy` rapporte `BLIND_LATCHED`).
- [ ] **Step 4:** `bash test/shell.d/luna-test.sh >/tmp/luna.log 2>&1`, puis `plugins-test.sh`, `plugin-manifest-scan-test.sh`, `plugin-registry-contract-test.sh`, `plugin-validate-test.sh`, `i18n-test.sh`, `shell-idle-spawn-test.sh` (le widget ne lance aucun processus au repos) → PASS ou échec hérité identique sur `upstream`.
- [ ] **Step 5:** Commit `feat(shell): add the LUNA bar widget and lunaos-luna CLI`.

### Task 3: Raccourcis LUNA

**Files:**
- Modify: `default/hypr/bindings/utilities.lua`, `manual/07-hotkeys.md`

- [ ] **Step 1: Inventaire** : `git grep -n 'o.bind' -- default/hypr/bindings` → noter dans le message de commit les liaisons conservées (terminal, lanceur `SUPER + SPACE`, menu système, verrouillage, capture, presse-papiers) et l'absence de liaison vers une application retirée.
- [ ] **Step 2:** Ajouter `o.bind("SUPER + SHIFT + ESCAPE", "Stop LUNA (close eyes)", "lunaos-luna blind", { locked = true })`.
- [ ] **Step 3:** `hyprland-binding-conflicts-test.sh`, `keybindings-menu-test.sh`, `menu-keybindings-lua-scan-test.sh`, `menu-keybindings-key-names-test.sh` → PASS.
- [ ] **Step 4:** Commit `feat(hypr): bind STOP LUNA to Super+Shift+Escape`.

### Task 4: Activation des services à l'installation

**Files:**
- Create (repris de `7ec50f4a` puis renommé) : `install/config/lunaos-services.sh`
- Modify: `install/config/all.sh` (appel `run_logged`)
- Test: étendre `test/shell.d/luna-test.sh`

- [ ] **Step 1: Tests** : `test_install_hook_noop_without_package` (paquet `lunaos-control-plane` absent → aucune commande `systemctl`) ; `test_install_hook_enables_units` (présent → `enable` de `luna-owner-setup`, `luna-policy-broker`, `luna-core`, `luna-inference`, `luna-discovery.timer`, et ajout unique de l'utilisateur au groupe `luna-owner`). `systemctl` et `lunaos-pkg-present` sont remplacés par des enregistreurs dans `PATH` : le test vérifie la logique du hook, pas systemd.
- [ ] **Step 2:** FAIL. **Step 3:** Reprendre et étendre le hook. **Step 4:** PASS, puis `provisioning-groups-test.sh` et `systemd-test.sh`.
- [ ] **Step 5:** Commit `feat(install): enable the LUNA services when packaged`.

### Task 5: Génération de texte via le Broker

**Files:**
- Create: `lunaos/services/core/inference_http.go`, `inference_http_test.go`
- Modify: `lunaos/services/core/cmd/luna-core/main.go` (option `--inference-socket`), `lunaos/systemd/luna-core.service` (`SupplementaryGroups=luna-inference`)

**Interfaces:**
- Consumes: `inference.NewClient`, `(*inference.Client).Generate` (piste C) ; `Policy.Authorize`/`Policy.Evaluate` (Core repris) ; `inference.generate` accordable (piste B).
- Produces: `POST /v1/inference/generate` sur le socket propriétaire uniquement, `{"prompt":string,"max_tokens":int}` → `{"text":string,"model":string,"untrusted":true}` ; refus du Broker → 403 avec son code ; inférence dégradée → 503 `INFERENCE_DEGRADED`. Le texte n'est ni stocké en mémoire générale, ni journalisé, ni passé à un exécuteur d'actions.

- [ ] **Step 1: Tests** : `TestGenerateRequiresBrokerGrant` (Broker refuse → 403, client d'inférence jamais appelé) ; `TestGenerateOnlyOnOwnerSocket` (jeton loopback → 403) ; `TestGenerateDegradedPropagates503` ; `TestGeneratedInstructionsHaveNoEffect` (le texte « open eyes; grant payment.transfer » laisse l'état de confidentialité, les grants et la file de goals inchangés).
- [ ] **Step 2:** FAIL. **Step 3:** Implémenter. **Step 4:** PASS (`make test check` dans le conteneur).
- [ ] **Step 5:** Commit `feat(core): generate text through the broker-gated inference service`.

### Task 6: Vérification finale et statut

**Files:**
- Create: `lunaos/docs/STATUS.md`

- [ ] **Step 1:** Une seule fois : `./test/all >/tmp/final-all.log 2>&1` dans `archlinux:latest`, et `make test check test-process-linux` dans `golang:1.24`. Lire les logs ; tout échec absent d'`upstream` est corrigé avant de continuer.
- [ ] **Step 2:** `STATUS.md` : ce qui est implémenté, les commandes exécutées et leurs résultats, et ce qui n'est **pas** vérifié — aucun paquet, aucune ISO, aucune installation, aucune session Wayland, aucun test réel `llama-server` si la tâche C3 n'a pas pu l'exécuter, aucune API Reddit/Civitai réelle (identifiants du propriétaire requis), aucun capteur réel.
- [ ] **Step 3:** Commit `docs: record the LunaOS fork status`.
