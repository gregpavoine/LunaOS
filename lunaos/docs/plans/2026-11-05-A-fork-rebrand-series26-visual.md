# Piste A — Fork : renommage, Series 26, visuel LUNA — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transformer l'arbre Omarchy en arbre LunaOS par un script rejouable, puis appliquer les choix Series 26 et le visuel LUNA par défaut.

**Architecture:** `lunaos/rebrand/rebrand.sh` réécrit chemins et contenus hors `lunaos/` en respectant `lunaos/rebrand/exceptions.txt`. Son résultat est committé seul (commit généré), puis les changements LunaOS viennent en commits normaux par-dessus.

**Tech Stack:** Bash 5, git, GNU sed/grep (dans un conteneur `archlinux:latest` si macOS), suites `./test/all`.

**Spec:** `lunaos/docs/design/2026-11-05-lunaos-fork-design.md` (§2)

## Global Constraints

- Possède tout ce qui est hors de `lunaos/`, plus `lunaos/rebrand/`. Ne touche à rien d'autre sous `lunaos/`.
- Style `AGENTS.md` : `#!/bin/bash`, deux espaces, `[[ ]]` / `(( ))`, `if`/`else` complet.
- Substitutions : `omarchy`→`lunaos`, `Omarchy`→`LunaOS`, `OMARCHY`→`LUNAOS`. Toute autre casse (ex. `OmArchy`) fait échouer le script.
- Exceptions minimales (avec raison dans `exceptions.txt`) : `linux-omarchy`, `linux-omarchy-headers`, `omarchy-nvim`, `omarchy-mac`, `omarchy-mac-boot`, `omarchy-mac-pkgs`, `omarchy-hyprland-titlebars`, `omarchy-keyring` s'il apparaît, `pkgs.omarchy.org`, `mirror.omarchy.org`, `stable-mirror.omarchy.org`, `rc-mirror.omarchy.org`, `arm-mirror.omarchy.org`, `omacom/omarchy-iso`, `omacom/omarchy-pkgs`, `omacom/omarchy-mac-pkgs`, `basecamp/omarchy` dans les crédits, le nom de famille interne de la police `omarchy.ttf`, `LICENSE`. Les noms `oma*` (omawrite, omasnap, omacut) ne contiennent pas « omarchy » et ne sont pas touchés.
- Dépôt de mise à jour : `https://github.com/omacom/omarchy.git` et `…/omacom/omarchy/releases…` deviennent `https://github.com/gregpavoine/LunaOS.git` / `…/gregpavoine/LunaOS/releases…` (substitution explicite, pas une exception).
- Fichiers binaires (`git grep -I` les exclut) : jamais réécrits en contenu ; seuls leurs chemins sont renommés.
- Installation neuve uniquement : aucune migration vers les anciens chemins `~/.config/omarchy`.
- Ne jamais pousser ; attribution `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

- La police d'icônes : le fichier est renommé mais sa famille interne reste `omarchy` ; les références QML/CSS à la famille doivent rester valides → test dans la tâche 1 (exception de famille) et vérification `fc-scan` dans la tâche 2.
- Un chemin contenant deux occurrences (`omarchy/omarchy-menu.jsonc`) ou un fichier et un répertoire renommés dans le même lot → test de fixture tâche 1.
- Un nom composé contenant une exception (`linux-omarchy-headers` contient `linux-omarchy`) → l'exception la plus longue doit gagner ; test tâche 1.
- `theme-set luna` puis passage à un autre thème et retour : l'arrière-plan noir revient → test tâche 4.
- Le lanceur et le dock ne doivent plus proposer une application retirée (entrées `.desktop`, dock épinglé) → test tâche 3.

---

### Task 1: Script de renommage et ses tests

**Files:**
- Create: `lunaos/rebrand/rebrand.sh`, `lunaos/rebrand/exceptions.txt`, `lunaos/rebrand/rebrand-test.sh`

**Interfaces:**
- Produces: `lunaos/rebrand/rebrand.sh <tree>` — réécrit `<tree>` en place (doit être un worktree git propre, sinon code 2) ; ignore `<tree>/lunaos/` et `.git/` ; imprime le nombre de chemins renommés et de fichiers modifiés ; code 3 si une casse inconnue est trouvée. `lunaos/rebrand/rebrand.sh --check <tree>` : code 0 si plus aucune occurrence hors exceptions, sinon liste `fichier:ligne` et code 1. Format `exceptions.txt` : une ligne `motif<TAB>raison`, `#` pour commentaire.

- [ ] **Step 1: Tests de fixture** dans `rebrand-test.sh` (sourcer `test/shell.d/base-test.sh` n'est pas possible ici : utiliser `set -euo pipefail` et des `assert_eq` locaux). Chaque test crée un dépôt git temporaire :
  - `test_renames_paths_and_contents` : `bin/omarchy-foo` contenant `OMARCHY_PATH Omarchy omarchy` → `bin/lunaos-foo` contenant `LUNAOS_PATH LunaOS lunaos`.
  - `test_nested_path` : `default/omarchy/omarchy-menu.jsonc` → `default/lunaos/lunaos-menu.jsonc`.
  - `test_longest_exception_wins` : `linux-omarchy-headers linux-omarchy omarchy-nvim` inchangés ; `omarchy-update` → `lunaos-update`.
  - `test_update_repo_rewritten` : `https://github.com/omacom/omarchy.git` → `https://github.com/gregpavoine/LunaOS.git` ; `omacom/omarchy-pkgs` inchangé.
  - `test_skips_lunaos_dir` : `lunaos/docs/x.md` contenant `omarchy` inchangé.
  - `test_binary_content_untouched` : un fichier binaire `default/fonts/omarchy/omarchy.ttf` (octets `\0omarchy\0`) → renommé en `default/fonts/lunaos/lunaos.ttf`, octets identiques.
  - `test_unknown_case_fails` : `OmArchy` → code 3, arbre inchangé.
  - `test_idempotent` : deux exécutions → `git status --porcelain` vide après commit de la première.
  - `test_check_mode` : après renommage, `--check` → code 0 ; avec une occurrence ajoutée → code 1 et la ligne listée.
- [ ] **Step 2:** `bash lunaos/rebrand/rebrand-test.sh >/tmp/rebrand.log 2>&1` → FAIL (script absent).
- [ ] **Step 3: Implémenter `rebrand.sh`.** Approche : (1) masquer chaque exception par un jeton unique `@@LUNAOS_KEEP_<n>@@` (exceptions triées par longueur décroissante), (2) appliquer les substitutions de dépôt, puis les trois casses, (3) restaurer les jetons ; chemins renommés du plus profond au moins profond avec `git mv` ; fichiers texte listés par `git grep -lI -i omarchy -- . ':!lunaos'`.
- [ ] **Step 4:** Relancer → PASS.
- [ ] **Step 5:** Commit `feat(rebrand): add the replayable Omarchy to LunaOS rename`.

### Task 2: Appliquer le renommage à l'arbre

**Files:** tout l'arbre hors `lunaos/` (généré).

- [ ] **Step 1: Inventaire** : `git grep -ohIiE '[a-z0-9.:/_@-]*omarchy[a-z0-9._/-]*' -- . ':!lunaos' | sort | uniq -c | sort -rn > /tmp/omarchy-tokens.txt`. Chaque jeton qui désigne une ressource externe (paquet, dépôt, miroir, URL) et n'est ni dans `exceptions.txt` ni dans les substitutions de dépôt est ajouté à `exceptions.txt` avec sa raison, et le test de la tâche 1 relancé. Commit séparé si `exceptions.txt` change.
- [ ] **Step 2:** `lunaos/rebrand/rebrand.sh .` puis `lunaos/rebrand/rebrand.sh --check .` → code 0.
- [ ] **Step 3:** `fc-scan --format '%{family}\n' default/fonts/lunaos/lunaos.ttf` → toujours la famille d'origine ; `git grep -n "font-family\|font.family" -- shell config default | grep -i omarchy` ne liste que des références à cette famille (exception).
- [ ] **Step 4:** `version` contient `4.0.0.alpha` : le remplacer par `0.1.0-dev` (LunaOS a sa propre numérotation). Logo `logo.txt`/`logo.svg`/`icon.*` : laissés tels quels (le visuel LUNA est en tâche 4).
- [ ] **Step 5: Suites, comparées à l'amont.** Dans un conteneur `archlinux:latest` (avec `base-devel git lua jq` installés) : `./test/all >/tmp/all-lunaos.log 2>&1` sur l'arbre renommé, et une fois `./test/all >/tmp/all-upstream.log 2>&1` sur un worktree `upstream`. Toute différence d'échec est corrigée (exception manquante ou substitution erronée), puis seulement le fichier de test concerné est relancé. Les échecs présents aussi sur `upstream` sont listés dans le commit, pas corrigés.
- [ ] **Step 6:** Commit `chore: rename Omarchy to LunaOS (generated by lunaos/rebrand/rebrand.sh)` — message indiquant le commit `upstream` source et la commande.

### Task 3: Series 26 — retraits d'applications

**Files:**
- Modify: `install/lunaos-base.packages` (retirer `libreoffice-fresh`, `moonlight-qt`, `omacalc`, `pinta`, et `kdenlive`/`omacut`/`xournalpp` s'ils y figurent), `default/hypr/bindings/applications.lua` (retirer les deux liaisons Calculator), `default/hypr/apps/system.lua` (règle `omacalc`), `default/hypr/apps/moonlight.lua` (supprimer) et son `require`, `config/lunaos/shell.json` (retirer `omacalc` du dock), `default/lunaos/launcher.hides` (retirer les lignes `libreoffice-*` devenues inutiles), `manual/07-hotkeys.md`, `manual/22-guis.md`, `manual/26-gaming.md` (anglais uniquement ; les traductions suivent le processus i18n existant, signalé dans le commit)
- Test: `test/shell.d/` — étendre le test existant qui valide `lunaos-pkg-defaults` (`git grep -l pkg-defaults test/`)

- [ ] **Step 1: Test** dans ce fichier existant : `test_series26_removed_packages_absent` — pour `x86` et `aarch64`, la sortie de `lunaos-pkg-defaults` ne contient aucun de `libreoffice-fresh kdenlive omacut pinta xournalpp moonlight-qt omacalc`, et contient toujours `obsidian obs-studio chromium localsend cliamp foot neovim docker`.
- [ ] **Step 2:** Lancer ce fichier → FAIL.
- [ ] **Step 3:** Appliquer les retraits listés. `bin/lunaos-remove-preinstalls` et `bin/lunaos-install-preinstalls` gardent leurs listes (outils optionnels de l'utilisateur).
- [ ] **Step 4:** Relancer ce fichier, plus `hyprland-binding-conflicts-test.sh`, `keybindings-menu-test.sh`, `menu-keybindings-lua-scan-test.sh`, `config-test.sh` → PASS.
- [ ] **Step 5:** Commit `feat: drop the Series 26 removed applications from the defaults`.

### Task 4: Visuel LUNA par défaut

**Files:**
- Create: `themes/luna/colors.toml` (copie de `themes/vantablack/colors.toml` ; `accent = "#8d8d8d"`), `themes/luna/backgrounds/0-black.png` (PNG 16×16 entièrement `#000000`, mis à l'échelle par le shell), `themes/luna/icons.theme` (copie de vantablack), `themes/luna/preview.png`
- Modify: `install/user/theme.sh` (`"Tokyo Night"` → `"Luna"` aux deux endroits), `config/lunaos/shell.json` (`bar.position` `"top"` → `"bottom"`)
- Test: étendre le test de thème existant (`git grep -l "theme-set" test/shell.d | head`)

- [ ] **Step 1: Tests** : `test_luna_theme_is_default` — exécuter `install/user/theme.sh` avec `HOME` temporaire et `LUNAOS_SETUP_CONTEXT=iso-chroot` → `~/.local/state/lunaos/current/theme.name` vaut `luna` ; `test_luna_background_is_black` — l'arrière-plan courant pointe vers `themes/luna/backgrounds/0-black.png` ; après `theme-set "Tokyo Night"` puis `theme-set Luna`, il pointe de nouveau vers ce fichier.
- [ ] **Step 2:** Lancer → FAIL.
- [ ] **Step 3:** Créer le thème et modifier les deux fichiers.
- [ ] **Step 4:** Relancer le fichier de test de thème, `background-intro-test.sh`, `background-pending-theme-test.sh` → PASS.
- [ ] **Step 5: Vérification visuelle** selon `agents/skills/visual-verification.md` si une session Hyprland est disponible ; sinon, le dire dans le commit (« non vérifié visuellement »).
- [ ] **Step 6:** Commit `feat(theme): make the black Luna theme and bottom bar the default`.
