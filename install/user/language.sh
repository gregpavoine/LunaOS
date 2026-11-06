# A language the install media preset (/etc/lunaos/language) becomes this
# user's desktop language; the system step generated its locale.
code=
[[ -r /etc/lunaos/language ]] && read -r code </etc/lunaos/language
if [[ -n $code ]]; then
  "$LUNAOS_PATH/bin/lunaos-language" seed "$code" || echo "Could not set the preset language $code" >&2
fi
