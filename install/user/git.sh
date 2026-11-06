# Set identification from install inputs
if [[ -n ${LUNAOS_USER_NAME//[[:space:]]/} ]]; then
  git config --global user.name "$LUNAOS_USER_NAME"
fi

if [[ -n ${LUNAOS_USER_EMAIL//[[:space:]]/} ]]; then
  git config --global user.email "$LUNAOS_USER_EMAIL"
fi
