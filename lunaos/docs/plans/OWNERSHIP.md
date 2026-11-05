# Propriété des fichiers entre les pistes

Ce document fixe qui peut modifier quoi pendant les quatre pistes parallèles (A fork, B cerveau, C LLM, D NSFW). Il reprend la règle stricte du design (`lunaos/docs/design/2026-11-05-lunaos-fork-design.md`, section « Propriété des fichiers ») et y ajoute les attributions fines décidées pendant la phase 0.

Ce fichier et tout ce qui se trouve sous `lunaos/docs/` (plans compris) appartiennent au contrôleur de la phase 0. Les pistes ne modifient ni les plans ni `OWNERSHIP.md`.

## Table de propriété

| Piste | Possède | Ne touche jamais |
|---|---|---|
| A — Fork | tout ce qui est hors de `lunaos/`, plus `lunaos/rebrand/` (seule exception sous `lunaos/`) | `lunaos/` hors `lunaos/rebrand/`, et les dossiers de B, C et D |
| B — Cerveau | `lunaos/services/{core,policy,identity}/` (sauf les fichiers de D listés plus bas), `lunaos/systemd/`, `lunaos/systemd-sysusers/`, `lunaos/tests/` | hors `lunaos/`, et les dossiers de A, C et D |
| C — LLM | `lunaos/services/inference/` (dont `systemd/` pour ses unités), `lunaos/models/` | hors `lunaos/`, et les dossiers de A, B et D |
| D — NSFW | `lunaos/services/intimacy/`, `lunaos/services/discovery/` (dont `systemd/` pour ses unités), et les fichiers de `lunaos/services/core/` listés plus bas | hors `lunaos/`, et les dossiers de A, B et C (sauf les fichiers de `lunaos/services/core/` listés plus bas) |

« Idem » ne s'applique pas entre pistes : C et D ne touchent ni les dossiers de B, ni ceux de l'autre piste (sauf les fichiers de `lunaos/services/core/` listés plus bas pour D).

La colonne « Ne touche jamais » s'entend hors fichiers partagés : les fichiers régis par la règle `shared:` (`lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile`, `lunaos/contracts/`) font exception pour toutes les pistes.

## Attributions fines

### Dans `lunaos/services/core/` (B et D)

D possède exactement ces fichiers, et leurs seuls tests :

- `lunaos/services/core/subjects_http.go` et `subjects_http_test.go`
- `lunaos/services/core/library_http.go` et `library_http_test.go`
- `lunaos/services/core/intimacy_context.go` et `intimacy_http_test.go`
- `lunaos/services/core/preference_http_test.go`
- `lunaos/services/core/discovery_policy_http_test.go`

Tout autre fichier de `lunaos/services/core/` appartient à B, même s'il référence `intimacy`. Cela inclut `http.go` et `http_test.go`.

### Unités systemd

- B possède `lunaos/systemd/` et `lunaos/systemd-sysusers/`.
- C met ses unités dans `lunaos/services/inference/systemd/`.
- D met ses unités dans `lunaos/services/discovery/systemd/`.

### Fichiers partagés

`lunaos/go.mod`, `lunaos/go.sum`, `lunaos/Makefile` et `lunaos/contracts/` n'ont pas de propriétaire unique. Toute piste peut les modifier, mais seulement dans un commit isolé dont le sujet commence par `shared:` et énonce le changement. Ce commit ne contient rien d'autre. `lunaos/contracts/actions.json` est la liste d'autorisation du Policy Broker : toute modification exige en plus la relecture de B ou du contrôleur.

`lunaos/services/internal/` est en lecture seule pour les pistes. Seul le contrôleur y modifie quelque chose.

## Chemins sans propriétaire de piste

Ces chemins n'appartiennent à aucune piste et sont tenus par le contrôleur de la phase 0 :

- `lunaos/docs/` et tout ce qu'il contient, plans compris ;
- `lunaos/README` et les autres fichiers à la racine de `lunaos/`, sauf les fichiers partagés ci-dessus (`go.mod`, `go.sum`, `Makefile`, `contracts/`) ;
- tout chemin sous `lunaos/` qui n'est listé dans aucune ligne du tableau ci-dessus, sauf les fichiers partagés ci-dessus (`go.mod`, `go.sum`, `Makefile`, `contracts/`).

## Faits partagés

- Branche d'intégration : `claude/omarchy-fork-specs-ad9dd4`. Chaque piste part de la tête de la branche d'intégration `claude/omarchy-fork-specs-ad9dd4` au moment du lancement des pistes, dans son propre worktree et sa propre branche.
- Branche locale `upstream` (créée par `git branch upstream`, ce n'est pas un remote) : Omarchy à `99f479b6`.
- Registre partagé : `contracts.Actions()` et `contracts.ConsentScopes()` sont exportés par `lunaos/contracts/`. Toute nouvelle action ou nouveau scope de consentement passe par ce registre, pas par une liste locale.
- Scopes intimes : `intimacy.KnownScopes()` est exporté par `lunaos/services/intimacy/subjects.go`.
- Identité du pair : le package `lunaos.local/lunaos/services/internal/peer` fournit `UID`, `ConnContext`, `FromContext` et `WithUID`. Ce package est en lecture seule pour les pistes ; les services qui authentifient un pair sur socket Unix l'utilisent tel quel.
- Module Go : un seul module, `lunaos.local/lunaos`.
- Docker Desktop ne répond pas sur ce Mac pour le moment. La vérification Go se fait donc nativement avec `cd lunaos && make test check`. Les tests réservés à Linux restent une dette enregistrée jusqu'à ce qu'une exécution en conteneur soit possible.

## Procédure en cas de besoin

Si une piste a besoin de modifier un fichier qu'elle ne possède pas, et qui n'est pas un fichier partagé ci-dessus, elle ne le modifie pas. Elle consigne la demande dans son plan, et le propriétaire ou le contrôleur fait le changement.
