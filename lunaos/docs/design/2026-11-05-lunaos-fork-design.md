# LunaOS — fork d'Omarchy repartant de zéro

**Date :** 2026-11-05
**Statut :** design validé par le propriétaire, non implémenté
**Base :** Omarchy `99f479b6` (branche `quattro`), worktree propre
**Sources des exigences (lues, non copiées) :** `lunaos/docs/` du checkout principal, branche `feature/lunaos-mac-image` — Mont Blanc 1.0 unifiée, Goal Engine, Series 26, ADR 0005, OPEN_DECISIONS. En cas de conflit, ce design prime pour l'organisation du fork ; les invariants de sécurité U1 de Mont Blanc priment sur tout.

## Décisions du propriétaire (2026-11-05)

1. On repart d'un Omarchy propre ; on reprend les specs comme source d'exigences et on réimporte **sélectivement, après revue** le code existant qui tient la route.
2. Renommage **complet** Omarchy → LunaOS (commandes, chemins, variables), réalisé de façon **scriptée et rejouable** pour pouvoir continuer à intégrer l'amont.
3. **Installation neuve uniquement** : les choix Series 26 deviennent des valeurs par défaut. Pas d'outil de conversion d'une installation Omarchy existante.
4. Le premier chantier inclut le fork, le visuel LUNA, les raccourcis LUNA, le cerveau LUNA (Go), le moteur LLM local et le NSFW (Discovery + Intimacy).
5. Les pistes avancent **en parallèle**.
6. Les specs restent dans le checkout principal ; seul ce design de synthèse entre dans le fork.

## 1. Organisation et parallélisme

### Phase 0 (séquentielle, avant les pistes)

- Créer la branche `upstream` = Omarchy `99f479b6`, jamais modifiée à la main.
- Committer ce design.
- Figer les contrats partagés dans `lunaos/contracts/` : OpenAPI du Core (repris de `contracts/openapi/local.json` et `native-v1.json` après revue), opérations typées du Policy Broker, liste des scopes de consentement U1.5, événements de confidentialité (BLIND, époque de capture). Toute modification ultérieure d'un contrat passe par un commit dédié sur la branche d'intégration et est annoncée aux autres pistes.

### Propriété des fichiers (règle stricte)

| Piste | Possède | Ne touche jamais |
|---|---|---|
| A — Fork | tout ce qui est hors de `lunaos/` | `lunaos/` |
| B — Cerveau | `lunaos/services/{core,policy,identity}`, `lunaos/systemd/`, `lunaos/systemd-sysusers/` | hors `lunaos/` |
| C — LLM | `lunaos/services/inference/`, `lunaos/models/` | idem |
| D — NSFW | `lunaos/services/{intimacy,discovery}/` | idem |

Le script de renommage n'entre jamais dans `lunaos/` (arborescence propre à LunaOS, qui peut citer Omarchy dans ses documents) : le renommage de A ne crée aucun conflit avec B, C, D. Chaque piste travaille dans son propre worktree et sa propre branche, partant de la branche d'intégration après la phase 0. `lunaos/go.mod` est commun : un seul module Go `lunaos.local/lunaos` ; un ajout de dépendance se fait en commit isolé pour limiter les conflits.

### Intégration (après A)

- Hook d'installation (`install/config/lunaos-services.sh` dans l'arborescence renommée) qui active les services LUNA quand le paquet de contrôle est présent.
- Widget LUNA de la barre (`shell/plugins/luna/`) parlant au Core par le socket propriétaire.
- Merge des quatre branches, puis `./test/all` et les suites Go une fois.

## 2. Piste A — le fork

### Renommage

- `lunaos/rebrand/rebrand.sh` : déterministe, idempotent, exécuté sur un arbre `upstream` propre. Il applique `git mv` sur les chemins contenant `omarchy` puis remplace `omarchy`/`Omarchy`/`OMARCHY` par `lunaos`/`LunaOS`/`LUNAOS` dans le contenu.
- `lunaos/rebrand/exceptions.txt` : motifs conservés tels quels, chacun avec sa raison. Au minimum : paquets externes (`linux-omarchy`, `linux-omarchy-headers`, `omarchy-nvim`, `omarchy-mac`, `omarchy-mac-boot`, `omarchy-hyprland-titlebars`, `omacut`, `omawrite`, `omasnap`), dépôts et miroirs (`pkgs.omarchy.org`, `*mirror.omarchy.org`), dépôts GitHub amont (`basecamp/omarchy`, `omacom/*`), licence MIT et crédits de DHH, nom de la police `omarchy.ttf` si son nom interne l'exige.
- Test `lunaos/rebrand/rebrand-test.sh` : échoue si une occurrence de « omarchy » subsiste hors exceptions et hors `lunaos/` ; échoue si deux exécutions successives ne donnent pas le même arbre.
- Synchronisation amont : avancer `upstream`, rejouer le script sur une copie, merger le résultat dans la branche LunaOS. Les commits LunaOS ne modifient jamais le résultat brut du script sans le documenter.
- `docs/`, `manual/` et `AGENTS.md` sont renommés comme le reste ; le manuel passe sous le nom LunaOS mais conserve l'attribution d'origine.

### Series 26 (valeurs par défaut d'une installation neuve)

- Retirés des listes de paquets, du menu, des raccourcis et des applications : LibreOffice, Kdenlive/Omacut, Pinta, Xournal++, Moonlight, Omacalc.
- Conservés tels quels : lanceur (Super+Space), Obsidian/Omawrite, OBS, Chromium, LocalSend, Cliamp, Neovim, Foot, Docker, outils disques, tous les points d'entrée IA, web apps (HEY, Basecamp, Discord, X, Maps, YouTube, Contacts, Messages, Photos, WhatsApp, Zoom), presse-papiers, notifications, capture/enregistrement/dictée.
- Remplacés maintenant : fond d'écran et thèmes (§ visuel), raccourcis (§ raccourcis).
- Remplacés plus tard (sous-projet bureau) : barre LUNA complète, verrouillage LUNA, interface de mise à jour LUNA (moteur pacman/migrations conservé).

### Visuel LUNA

- Thème `themes/luna/` par défaut : fond `#000`, aucun fond d'écran, aucune icône de bureau, couleurs d'accent sobres.
- Barre en bas, texte seul (horloge, espaces de travail) ; statut LUNA discret dans un coin (via le widget d'intégration).
- Pas de photo de personne comme interface par défaut.

### Raccourcis LUNA

- Inventaire des liaisons existantes et des conflits, consigné dans le commit.
- Ajout d'un raccourci STOP LUNA (BLIND immédiat) accessible au clavier en toutes circonstances.
- Conservés : terminal, lanceur, secours, verrouillage, capture, presse-papiers.
- Retirés : raccourcis des applications supprimées.

## 3. Piste B — le cerveau LUNA

Reprise après revue depuis `feature/lunaos-mac-image` :

- **Core** (`luna-core`) : registre de goals SQLite, autorisation par le Broker **avant** le score puis revalidation avant exécution, API idempotente, evidence persistante, scheduler qui reprend après redémarrage, essais de stratégie bornés, contrôleur de confidentialité BLIND (démarre `STARTUP_LOCKED`, verrou persistant, sortie seulement par clic sur le socket propriétaire, époque de capture).
- **Policy Broker** (`luna-policy-broker`) : processus séparé, refus par défaut, grants explicites et scopés, budgets à zéro, vérification de l'UID du pair sur socket Unix, opérations typées uniquement (jamais de chaîne shell produite par un modèle).
- **Owner setup** (`luna-owner-setup`) : identifiants séparés propriétaire / Core↔Broker, jamais écrasés, jeton révélé seulement sur `--reveal-token`.
- systemd : un utilisateur dédié par service, sockets restreints, sysusers.
- **Non repris :** le checkpoint Rust, le code Python de référence, le shell GTK4, l'ancien profil Archiso imbriqué.
- Critère de reprise : chaque fichier importé est relu ; ses tests passent dans un conteneur Linux ; tout comportement non couvert par un test est soit testé, soit laissé de côté.

## 4. Piste C — moteur LLM local

- `luna-inference` : superviseur Go de **llama.cpp `llama-server`** (CUDA sur la RTX 3080 Ti 12 Go, Vulkan ou CPU sur Mac aarch64). Il démarre, surveille et arrête le serveur, et l'expose uniquement sur un socket Unix local.
- `lunaos/models/manifest.json` (JSON plutôt que TOML : pas de dépendance hors stdlib Go) : chaque modèle déclare URL, sha256, licence, taille, budget VRAM. Aucun fichier de modèle dans le dépôt ; téléchargement vérifié par hash, refus sinon.
- Accès : le Core appelle l'inférence uniquement avec un grant Broker `inference.generate`. **La sortie du modèle est une donnée non fiable** : elle ne peut ni autoriser une action, ni lever BLIND, ni élargir un scope.
- Politique VRAM initiale : ComfyUI prioritaire ; l'inférence se décharge à la demande. Pas de repli cloud (désactivé, non approuvé).
- BLIND n'arrête pas le texte local mais bloque toute entrée perceptive (caméra, micro, écran) vers le modèle.

## 5. Piste D — NSFW

### Intimacy (reprise après revue)

- Vault chiffré séparé de la mémoire générale ; sujets avec consentement par personne et par scope (liste U1.5), preuve par clic du sujet ou défi vocal à usage unique ; refus du consentement par procuration ; scopes intimes réservés aux adultes attestés ; `unverified`/`suspected_minor`/`minor` n'obtiennent rien et le signalement efface les actifs intimes ; révocation qui efface dans la même transaction avec reçu ; clé par sujet détruite à la suppression ; audit append-only sans contenu intime.
- Médiathèque privée et contexte intime : répond uniquement en contexte privé, intimacy non arrêtée ; seule l'UI de confiance peut ouvrir le contexte ou lever STOP_INTIMACY.
- Sauvegarde/restauration `luna-vault` et rotation de clé.

### Discovery (nouveau)

- Service `luna-discovery`, identité de service distincte du vault, clés de stockage distinctes.
- Connecteurs officiels uniquement : **Reddit** (client OAuth enregistré, accès API approuvé, en-têtes de quota respectés) et **Civitai** (clé API, liste blanche de catégories/tags/créateurs/niveau de contenu, pagination par curseur, **aucun crawl NSFW universel**). Pas de scraping, pas de mutualisation de comptes.
- Pipeline : ingestion incrémentale → provenance → contrôle des droits → métadonnées normalisées → téléchargement borné et reprenable → stockage local chiffré → dédoublonnage SHA-256 → index → flux.
- Fiabilité : `Retry-After`, backoff exponentiel, file des échecs, points de reprise ; un compte Reddit déconnecté ne casse ni Civitai ni la bibliothèque hors ligne. Respect des retraits du fournisseur et des licences.
- Les actifs du profil corporel ne sont **jamais** dans l'index Discovery ni dans les recommandations.
- Stockage : SQLite chiffré comme le reste. PostgreSQL + pgvector (demandé par la spec) est reporté jusqu'au besoin réel d'un index vectoriel ; décision consignée en ADR à ce moment-là.

### Invariants testés (non négociables, Mont Blanc U1)

Aucune capture ni recommandation impliquant un mineur ou un adulte non vérifié ; pas de consentement par procuration ; BLIND prime sur LUNA, les agents et le code généré ; aucune donnée sensible (média, embedding, miniature, poids) ne sort de la machine ; contenu externe licite, adulte, autorisé par le fournisseur ; aucune usurpation sexuelle d'une personne réelle identifiable.

## 6. Tests, vérification et limites

- A : `lunaos/rebrand/rebrand-test.sh`, puis `./test/all` sur l'arbre renommé ; tout échec présent aussi sur `upstream` est signalé, pas corrigé dans ce chantier.
- B, C, D : `go test ./...` et `go test -race` dans un conteneur Linux x86_64 ; test de processus réel Core/Broker ; tests d'invariants de confidentialité pour D ; pour C, un test contre un vrai `llama-server` avec un petit modèle épinglé.
- Intégration : merge, `./test/all` et suites Go une seule fois.
- **Hors périmètre, assumé :** paquets signés, ISO, Try Luna, verrouillage et interface de mise à jour LUNA, capteurs réels (caméra, micro, écran), ComfyUI, iPhone Sync, NAS. Sans paquets ni ISO, **le fork ne s'installe pas encore sur une machine** ; l'interface n'est pas vérifiée sous Wayland. Le statut le dit explicitement ; aucune fonctionnalité n'est déclarée « opérationnelle » sans test exécuté.
