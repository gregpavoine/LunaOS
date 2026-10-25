# Démarrage de LunaOS sur Mac M2 avec Codex

## Préparation
- MacBook Air M2 utilisé comme poste de pilotage, Linux x86_64 + NVIDIA comme machine cible Omarchy. Prévoir connexion réseau, accès SSH, Git, et assez d'espace disque sur le PC cible. L'utilisation du Mac pour les tests Python/API est possible; CUDA, Hyprland et systemd seront vérifiés sur Linux.
- Installer les Apple Command Line Tools: `xcode-select --install` si absents.
- Installer Codex CLI avec l'installateur officiel: `curl -fsSL https://chatgpt.com/codex/install.sh | sh` (inspecter le script téléchargé avant exécution si politique restrictive); alternative `npm install -g @openai/codex` si Node/npm déjà installés.
- `codex --version`, puis lancer `codex` et se connecter à ChatGPT. Suivre le menu interactif; `/status`, `/permissions`, `/model`, `/init` existent dans CLI récent. Ne pas présumer d'un niveau d'abonnement ou de droits d'exécution.

## Étapes exactes
1. Télécharger le pack ZIP LunaOS et le placer dans `~/Downloads`.
2. Dans Terminal macOS :

```bash
mkdir -p ~/Projects/lunaos
cd ~/Projects/lunaos
unzip -o ~/Downloads/LUNAOS_CODEX_MAC_START_PACK.zip
# Le ZIP inclut AGENTS.md et docs/ ; ne pas copier dans une installation Arch en production

git init
git add AGENTS.md docs README.md
git commit -m "docs: bootstrap LunaOS product requirements"
codex
```

3. Dans Codex, ouvrir `docs/FIRST_PROMPT_FOR_CODEX.md`, copier-coller la tâche initiale. Donner à Codex l'accès en écriture **au dépôt seulement**, et pas à l'ensemble du Mac. Vérifier et autoriser les commandes requises individuellement.
4. Après la première génération, vérifier `git diff`, `git status`, suites de tests, rapports d'implémentation. Demander une preuve d'exécution, pas une affirmation.
5. Préparer l'accès au Linux cible. **Sur la machine Linux elle-même**, vérifier que SSH est installé/activé par un administrateur; établir une clé depuis le Mac et ajouter la clé publique sur la cible, idéalement avec un utilisateur de déploiement limité.
6. Sur Mac : `ssh -o BatchMode=yes utilisateur@hote 'uname -a; cat /etc/os-release; command -v pacman; command -v nvidia-smi; nvidia-smi -L'` (remplacer valeurs réelles). Si SSH inaccessible, demander à Codex de continuer les composants portables et de marquer la validation Linux BLOQUÉE.
7. Transmettre à Codex uniquement un alias SSH/config dédié et non les secrets privés, puis lui demander l'inventaire **read-only** de l'Omarchy réel avant toute adaptation.
8. Ne pas lancer un nettoyage Omarchy, une conversion système, ni une mise à jour globale avant vérification des snapshots/sauvegardes et d'un accès de secours. Le script de nettoyage Series 26 doit rester en `--dry-run` jusqu'au feu vert.
9. Quand les premières fonctions sont stables, créer CI, sauvegardes, déploiement incrémental, et seulement ensuite des sessions automatisées longues (`codex exec`) sous politiques strictes et quotas. Un terminal ouvert 24/7 ne garantit pas la continuité si l'accès expire ou si des approbations sont nécessaires.

## Première preuve d'utilisation
- V0 : service de Goal Engine réellement persistant; action sûre exécutée; redémarrage/reprise; LLM local connecté; bureau Omarchy préservé; ComfyUI contrôlé sur RTX 3080 Ti avec artefact vérifiable.
- V1 : ajout progressif de tous les modules Mont Blanc, y compris Intimacy/consent/BLIND, iPhone Sync et Discovery, avec données de test et conformité. Ne jamais annoncer "terminé" sur la base d'un commit ou de tests fictifs.

## Référence officielle Codex CLI
- https://learn.chatgpt.com/docs/cli
- https://github.com/openai/codex
