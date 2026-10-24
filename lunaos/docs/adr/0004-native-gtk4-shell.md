# ADR 0004: GTK4 with gtk-rs for the native shell

Accepted 2026-10-24. GTK4 is selected over Slint for the Linux-first LunaOS shell because GTK is a mature desktop toolkit with strong Linux accessibility and integration conventions, and Arch packages the native development/runtime dependencies. The official gtk-rs guide identifies GTK4/Libadwaita and build tooling for Arch; gtk-rs documents maintained safe Rust bindings. GTK work remains on the GLib main thread, with async service responses delivered back to that context.

Slint remains a credible alternative with its portable Rust UI language and Winit/Wayland backends. Reconsider if LunaOS develops a non-Linux shell/embedded product requirement or if GTK integration does not pass the target accessibility and performance acceptance tests. GTK shell compilation and Wayland runtime have not yet been verified on Arch.
