require("default.hypr.startup-cursor")

hl.on("hyprland.start", function()
  -- Applications inherit the user's cursor; only the compositor starts blank.
  if lunaos_startup_cursor_pending then
    hl.env("XCURSOR_PATH", lunaos_startup_cursor.path)
    hl.env("XCURSOR_THEME", lunaos_startup_cursor.xcursor)
  end

  -- Slow app launch fix -- set systemd and D-Bus activation vars before
  -- starting session services.
  hl.exec_cmd("dbus-update-activation-environment --systemd --all")

  hl.exec_cmd("lunaos-launch-shell")
  hl.exec_cmd("lunaos-provision-first-run")
  hl.exec_cmd("lunaos-powerprofiles-init")
  hl.exec_cmd(o.launch("lunaos-hyprland-monitor-watch"))
  hl.exec_cmd(o.launch("udiskie --automount --no-notify --no-tray"))

  -- Run post-boot hooks after startup config has loaded.
  hl.exec_cmd("sleep 2 && lunaos-hook post-boot")
end)
