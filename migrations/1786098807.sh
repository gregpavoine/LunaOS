echo "Relink agent skill symlinks to default/agents/skills/lunaos"

mkdir -p ~/.agents/skills ~/.claude/skills ~/.codex/skills ~/.pi/agent/skills
ln -sfn "$LUNAOS_PATH/default/agents/skills/lunaos" ~/.agents/skills/lunaos
ln -sfn "$LUNAOS_PATH/default/agents/skills/lunaos" ~/.claude/skills/lunaos
ln -sfn "$LUNAOS_PATH/default/agents/skills/lunaos" ~/.codex/skills/lunaos
ln -sfn "$LUNAOS_PATH/default/agents/skills/lunaos" ~/.pi/agent/skills/lunaos
