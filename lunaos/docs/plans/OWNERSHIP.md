# Propriété des fichiers entre les pistes

Ce document fixe qui peut modifier quoi pendant les quatre pistes parallèles (A fork, B cerveau, C LLM, D NSFW). Il reprend la règle stricte du design (`lunaos/docs/design/2026-11-05-lunaos-fork-design.md`, section « Propriété des fichiers ») et y ajoute les attributions fines décidées pendant la phase 0.

## Table de propriété

| Piste | Possède | Ne touche jamais |
|---|---|---|
| A — Fork | tout ce qui est hors de `lunaos/` | `lunaos/` |
| B — Cerveau | `lunaos/services/{core,policy,identity}`, `lunaos/systemd/`, `lunaos/systemd-sysusers/` | hors `lunaos/` |
| C — LLM | `lunaos/services/inference/`, `lunaos/models/` | idem |
| D — NSFW | `lunaos/services/{intimacy,discovery}/` | idem |

## Attributions fines

- D possède aussi `lunaos/services/core/{subjects_http,library_http,intimacy_context}.go` et leurs tests : `subjects_http_test`, `library_http_test`, `intimacy_http_test`, `preference_http_test`, `discovery_policy_http_test`.
- B possède le reste de `lunaos/services/core/` (dont `http.go` et `http_test.go`).
- `lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile` et `lunaos/contracts/` ne changent que par commit isolé, annoncé aux autres pistes dans leur description de commit.
- `lunaos/services/internal/` appartient à la phase 0 : tout changement y est un commit isolé, annoncé aux autres pistes.
- Le script de renommage (`lunaos/rebrand/`) appartient à A. Il n'entre jamais dans `lunaos/` au-delà de son propre dossier, et son renommage ne crée aucun conflit avec B, C ou D.

## Faits partagés

Ces points sont établis par la phase 0 et valent pour toutes les pistes.

- Branche d'intégration : `claude/omarchy-fork-specs-ad9dd4`. Chaque piste part de ce commit dans son propre worktree et sa propre branche.
- Remote `upstream` : Omarchy à `99f479b6`.
- Registre partagé : `contracts.Actions()` et `contracts.ConsentScopes()` sont exportés par `lunaos/contracts/`. Toute nouvelle action ou nouveau scope de consentement passe par ce registre, pas par une liste locale.
- Scopes intimes : `intimacy.KnownScopes()` est exporté par `lunaos/services/intimacy/subjects.go`.
- Identité du pair : le package `lunaos.local/lunaos/services/internal/peer` fournit `UID`, `ConnContext`, `FromContext` et `WithUID`. Les services qui authentifient un pair sur socket Unix l'utilisent tel quel.
- Module Go : un seul module, `lunaos.local/lunaos`.
- Docker Desktop ne répond pas sur ce Mac pour le moment. La vérification Go se fait donc nativement avec `cd lunaos && make test check`. Les tests réservés à Linux restent une dette enregistrée jusqu'à ce qu'une exécution en conteneur soit possible.

## Procédure en cas de besoin

Si une piste doit toucher un fichier qu'elle ne possède pas, elle ne le modifie pas. Elle ouvre une demande dans son plan, et le propriétaire ou la phase 0 fait le changement dans un commit isolé.
