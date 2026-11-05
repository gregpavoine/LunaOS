# Piste B — Cerveau LUNA (Core, Policy Broker, identité) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Faire du code repris en phase 0 un plan de contrôle prêt à accueillir l'inférence : Broker piloté par le registre d'actions, unités systemd durcies et vérifiées.

**Architecture:** Le Broker remplace sa liste en dur (`local.checksum` seul) par `contracts.Actions()`. Les règles fixes (paiement refusé, classe de confidentialité `general` seule, grant HMAC lié au goal et à l'action, une heure) restent inchangées. Les unités systemd sont durcies et passées à `systemd-analyze verify` dans un conteneur Arch.

**Tech Stack:** Go 1.24, SQLite, systemd, conteneurs `golang:1.24` et `archlinux:latest`.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md` (§3)

## Global Constraints

- Possède `lunaos/services/{core,policy,identity}/` (sauf les fichiers Core attribués à D dans `OWNERSHIP.md`), `lunaos/systemd/`, `lunaos/systemd-sysusers/`.
- Refus par défaut ; budgets à zéro ; `payment.*` → `BUDGET_UNSET` ; `privacy_class` ≠ `general` → `PRIVACY_SCOPE_DENIED` ; aucune opération ne prend une chaîne shell.
- Tests Go sous Linux uniquement (`docker run --rm -v "$PWD/lunaos":/src -w /src golang:1.24 make test check`).
- Ne jamais pousser ; attribution `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- Une action listée dans le registre mais sans exécuteur dans le Core : le Broker l'autorise, le Core doit répondre `ACTION_NOT_ALLOWED` sans plantage → test tâche 1.
- Un grant valide pour `local.checksum` rejoué pour `inference.generate` → refusé → test tâche 1.
- Un grant expiré d'une seconde → refusé → test tâche 1.
- Un durcissement systemd qui empêche le Core d'écrire sa base (`StateDirectory`) ou d'ouvrir le socket du Broker → le test de processus réel doit passer avec les unités → tâche 2.
- Redémarrage du Broker pendant qu'un goal est en file : le goal reste en attente puis reprend → déjà couvert par le test de processus importé ; le relancer en tâche 2.

---

### Task 1: Broker piloté par le registre d'actions

**Files:**
- Modify: `lunaos/services/policy/main.go`
- Test: `lunaos/services/policy/main_test.go`

**Interfaces:**
- Consumes: `contracts.Actions() []contracts.Action` (phase 0).
- Produces: le Broker accorde un grant `OWNER_SCOPED_GRANT` pour toute action du registre dont `privacy_class` correspond à la requête ; codes inchangés pour le reste. `inference.generate` devient accordable (utilisé par l'intégration Z5).

- [ ] **Step 1: Tests** : `TestGrantForRegisteredInferenceAction` (`inference.generate`, `general` → `Allowed`, `OWNER_SCOPED_GRANT`) ; `TestUnregisteredActionDenied` (`shell.exec` → `ACTION_DENIED`) ; `TestGrantNotReusableAcrossActions` (grant `local.checksum` évalué pour `inference.generate` → `GRANT_INVALID_OR_EXPIRED`) ; `TestExpiredGrantDenied` (grant signé avec `Expires = now-1`) ; `TestPaymentStillUnset` ; `TestPrivacyClassMismatchDenied` (`inference.generate` + `sensitive` → `PRIVACY_SCOPE_DENIED`).
- [ ] **Step 2:** Lancer dans le conteneur → FAIL sur les deux premiers.
- [ ] **Step 3:** Construire au démarrage `map[string]string` nom→classe depuis `contracts.Actions()` ; remplacer le test `req.Action != "local.checksum"` par une recherche dans cette table.
  Valider le registre au démarrage du Broker (noms vides ou en double, noms `payment.`, classes de confidentialité inconnues rejetés) et refuser de démarrer en cas d'échec, avec le test `TestBrokerRefusesInvalidRegistry`.
- [ ] **Step 4:** Test Core `TestRegisteredActionWithoutRunnerIsRejected` dans `lunaos/services/core/engine_test.go` : un goal `inference.generate` approuvé par une politique de test se termine en échec `ACTION_NOT_ALLOWED`, sans panique, evidence non vérifiée.
- [ ] **Step 5:** Relancer `make test check` → PASS.
- [ ] **Step 6:** Commit `feat(policy): grant actions from the shared registry`.

### Task 2: Unités systemd durcies et vérifiées

**Files:**
- Modify: `lunaos/systemd/luna-core.service`, `luna-policy-broker.service`, `luna-owner-setup.service`
- Create: `lunaos/tests/systemd-units-test.sh`

- [ ] **Step 1: Test** `lunaos/tests/systemd-units-test.sh` (exécuté dans `archlinux:latest`) : `systemd-analyze verify` sur les trois unités (avec des binaires factices sous `/usr/bin` créés par le test, uniquement pour satisfaire `ExecStart`) → aucune erreur ; chaque unité de service long contient `NoNewPrivileges=yes`, `ProtectSystem=strict`, `ProtectHome=yes`, `PrivateTmp=yes`, `RestrictAddressFamilies=AF_UNIX` (le Core ajoute `AF_INET` seulement s'il écoute en loopback), `User=` dédié (`lunaos` / `luna-policy`) ; `systemd-analyze security <unit>` imprimé dans le log (score noté, pas d'assertion).
- [ ] **Step 2:** Lancer → FAIL.
- [ ] **Step 3:** Ajouter les directives ; `StateDirectory=`/`RuntimeDirectory=` pour les chemins que les services écrivent ; `Restart=on-failure` avec `StartLimitBurst=5`, `StartLimitIntervalSec=60`.
- [ ] **Step 4:** Relancer le test ; relancer `make test-process-linux` (test de processus réel importé) → PASS.
- [ ] **Step 5:** Ajouter `bash tests/systemd-units-test.sh` à la cible `check` du Makefile (commit isolé annoncé, cf. `OWNERSHIP.md`).
- [ ] **Step 6:** Commit `feat(systemd): harden and verify the LUNA service units`.

### Task 3: Rafraîchir le contrat OpenAPI du Core

**Files:**
- Modify: `lunaos/contracts/openapi/local.json` (commit `shared:`)
- Test: un test Go dans `lunaos/contracts/` ou `lunaos/services/core/`

- [ ] **Step 1: Test** : étendre `make check` avec un test Go qui vérifie que chaque route enregistrée par le handler HTTP du Core apparaît dans `local.json`.
- [ ] **Step 2:** Lancer → FAIL (routes manquantes).
- [ ] **Step 3:** Ajouter à `local.json` les routes `/v1/privacy/{blind,open-eyes,status,leases,leases/validate}`, `/v1/subjects*`, `/v1/intimacy/library/*`, `/v1/intimacy/context` et `/v1/intimacy/resume`, avec les formes de requête et de réponse prises des handlers du Core.
- [ ] **Step 4:** Relancer `make test check` → PASS.
- [ ] **Step 5:** Commit `shared: document the Core privacy, subject and intimacy routes`.
