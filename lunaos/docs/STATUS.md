# État de LunaOS

Ce document dit ce qui est implémenté, ce qui a été exécuté et avec quel résultat, et ce qui n'est **pas** vérifié. Il est à jour à la date de la dernière entrée de la section « Vérifications ».

## Implémenté

- **Cerveau LUNA (Go)** : `luna-core` (goals, scheduler, routes privacy, subjects, intimacy, discovery, génération de texte), `luna-policy-broker` (refus par défaut, budgets à zéro, `payment.*` refusé, classe `general` seule), `luna-owner-setup` (identité).
- **Route de génération** : `POST /v1/inference/generate` sur le socket propriétaire uniquement. Elle passe par le Broker (Authorize puis Evaluate) pour l'action `inference.generate`, appelle le service d'inférence, et renvoie le texte avec `untrusted: true`. Le texte n'est ni stocké, ni journalisé, ni interprété comme une action.
- **Inférence locale (C)** : `luna-inference` supervise `llama-server` en loopback, n'utilise que des modèles dont le SHA-256 correspond au manifeste, et répond `INFERENCE_DEGRADED` au lieu d'un texte vide.
- **Intimacy et discovery (D)** : vault chiffré AES-GCM, consentement par personne et par scope (13 scopes), connecteurs Reddit et Civitai avec `Retry-After`. Aucun appel réseau réel n'est fait dans les tests.
- **Fork (A)** : script de renommage `lunaos/rebrand/rebrand.sh`, tests de ce script, exceptions documentées, visuels Series 26.
- **Intégration bureau** :
  - CLI `lunaos-luna` (statut, goals, checksum, pause, reprise, annulation, privacy, blind, open-eyes). Sans Core, il sort avec le code 3.
  - Widget de barre `lunaos.luna` (plugin `shell/plugins/luna/`), ajouté à la section droite de la barre.
  - Raccourci `SUPER + SHIFT + ESCAPE` : « Stop LUNA » lance `lunaos-luna blind` avec `locked = true`.
  - Hook d'installation `install/config/lunaos-services.sh`, appelé depuis `install/config/all.sh` : il ne fait rien sans le paquet `lunaos-control-plane`.
- **Systemd** : `luna-core` dépend de `luna-inference` (socket d'inférence et groupe `luna-inference`). Les unités sont durcies (`NoNewPrivileges`, `ProtectSystem=strict`, utilisateurs dédiés).

## Vérifications exécutées

Environnement : macOS (Darwin, arm64), Go 1.24.13 local au projet, bash 3.2 du système.

| Commande | Résultat |
|---|---|
| `make -C lunaos test check` (Go 1.24.13) | OK |
| `GOOS=linux go vet ./...` | OK |
| `lunaos/rebrand/rebrand.sh --check .` | OK (code 0) |
| `lunaos/rebrand/rebrand-test.sh` | OK |
| `go test -count=3 -race ./services/... ./contracts/...` (Go 1.24.13, macOS) | OK, trois passes sans échec ni course détectée |
| `bash test/shell.d/luna-test.sh` | OK pour les cas sans Core réel (CLI sans Core, validation d'id, hook d'installation) ; la partie Core réel est sautée hors Linux |
| `bash test/shell.d/{plugins,plugin-manifest-scan,plugin-registry-contract,plugin-validate,provisioning-groups,systemd}-test.sh` | OK |

Échecs observés, **dus à l'environnement** et non au code modifié :

- `hyprland-binding-conflicts`, `keybindings-menu`, `menu-keybindings-lua-scan` : la commande `lua` n'est pas installée sur ce Mac.
- `menu-keybindings-key-names`, `test/cli`, `bin/lunaos` : bash 3.2 ne connaît pas `declare -A`.
- `i18n-test` : `wc` de BSD ajoute des espaces dans le comptage des langues.
- `shell-idle-spawn-test` : exige une session de bureau Wayland.

## Non vérifié

- **Suites Linux** : `make test-process-linux`, `./test/all` et les tests shell ci-dessus, à relancer dans `archlinux:latest` ou sur une machine Linux avec bash 5.
- **`systemd-analyze verify`** sur les unités : hors conteneur Linux, il est sauté.
- **Test réel `llama-server`** : il exige un modèle du manifeste et un paquet llama.cpp pour la plateforme. Non exécuté.
- **Reddit et Civitai** : aucun appel réel. Il faudrait des identifiants du propriétaire.
- **Capteurs réels** et **session Wayland** : non testés.
- **Installation système** : aucun paquet `lunaos-control-plane`, aucune ISO, aucune installation. Les binaires `luna-*`, le manifeste et les unités ne sont pas encore déployés par une étape d'installation, donc `lunaos-services.sh` n'a jamais été exécuté sur une vraie installation.
- **Traductions** : les 29 clés du widget sont en anglais seulement ; le catalogue français n'est pas encore complété.
