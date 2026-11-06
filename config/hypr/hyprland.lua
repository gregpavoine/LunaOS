-- Learn how to configure Hyprland: https://wiki.hypr.land/Configuring/Start/

-- LunaOS's bootstrap keeps path setup out of this user config.
dofile((os.getenv("LUNAOS_PATH") or "/usr/share/lunaos") .. "/default/hypr/bootstrap.lua")

-- Disable all LunaOS default bindings. Add your own in hypr/bindings.lua.
-- lunaos_default_bindings = false
--
-- Or disable only bindings for LunaOS's preinstalled apps/web apps while
-- keeping core window-manager bindings:
-- lunaos_preinstalled_bindings = false

-- Load LunaOS defaults.
require("default.hypr.lunaos")

-- Put your personal overrides in these files. They're loaded after LunaOS's
-- defaults so package updates can improve the defaults without rewriting your
-- ~/.config/hypr files.
require("hypr.monitors")
require("hypr.input")
require("hypr.bindings")
require("hypr.looknfeel")
require("hypr.autostart")

-- Toggle config flags dynamically.
require("default.hypr.toggles")

-- Add any other personal Hyprland configuration below.
-- o.window("qemu", { workspace = "5" })
