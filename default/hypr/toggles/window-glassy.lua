-- Frosted glass from the glassy Starship (#14726): whatever is behind a
-- see-through window or shell surface is blurred. Terminals get the see-through
-- background itself from lunaos-theme-set-glass, which keeps their text crisp.
require("default.hypr.glass")()

-- GTK apps fade as a whole, since GTK draws their backgrounds solid, and get
-- their blur from the backdrop-filter Glassy adds to the user's gtk.css.
o.window("org\\.gnome\\.Nautilus", { opacity = "0.75 0.75" })

hl.layer_rule({
  match = { namespace = "^(lunaos-bar|lunaos-menu|lunaos-notifications|lunaos-osd|lunaos-polkit|lunaos-reminders|lunaos-clipboard|lunaos-emojis|lunaos-image-selector|lunaos-keyboard-panel|lunaos-network-qr)$" },
  blur = true,
  ignore_alpha = 0.3,
})
